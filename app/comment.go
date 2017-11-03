// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"math"
	"strconv"
	"time"

	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

// Comment is a comment on a post.  A comment is associated to one post, and may have a parent comment
// so comments can be nested conversations
type Comment struct {
	Key      data.UUID `json:"key,omitempty" gorethink:",omitempty"`
	PostKey  data.UUID `json:"postKey,omitempty" gorethink:",omitempty"`
	Parent   data.UUID `json:"parent,omitempty" gorethink:",omitempty"`
	Username data.Key  `json:"username,omitempty" gorethink:",omitempty"`
	Comment  string    `json:"comment,omitempty" gorethink:",omitempty"`
	data.Version

	post   *Post
	parent *Comment
	user   *User
}

// CommentTree is a tree of comments including their children and a "HasChildren" tag if children exist that
// weren't retrieved.
type CommentTree struct {
	Comment
	Children    []CommentTree `json:"children,omitempty" gorethink:",omitempty"`
	HasChildren bool          `json:"hasChildren,omitempty" gorethink:",omitempty"`
	More        bool          `json:"more,omitempty" gorethink:",omitempty"` // more children exist
}

const (
	commentMaxRetrieve  = 500              //max # of comments that can be requested at a time
	commentEditDuration = 60 * time.Second // max amount of time before a comment can't be edited anymore
	commentMaxLength    = 50000
)

var (
	// ErrCommentNoUser is the error returned when no use is specified
	ErrCommentNoUser = fail.New("You must be logged in to comment.")
	// ErrCommentEmpty is when there is no comment text
	ErrCommentEmpty = fail.New("Your comment must contain some text")
	// ErrCommentTooLong is when the comment is too long
	ErrCommentTooLong = fail.New("Your comment is too long. The max length is " + strconv.Itoa(commentMaxLength))
	// ErrCommentNotAllowed is when comments aren't allowed on a specific post
	ErrCommentNotAllowed = fail.New("Commenting is not allowed on this post.")
	// ErrCommentPostNotPublished is when a comment has been applied to a non-published post
	ErrCommentPostNotPublished = fail.New("This post is not published, and cannot be commented on.")
)

// setMoreChildren recursively runs through the comment tree to check if there are more children (past the limit)
// that can be retrieved
func (t *CommentTree) setMoreChildren(limit int) {
	if len(t.Children) > limit {
		t.More = true
		t.Children = t.Children[:len(t.Children)-1]
	}
	for i := range t.Children {
		t.Children[i].setMoreChildren(limit)
	}
}

// CommentGet retrieves  a single comment
func CommentGet(key data.UUID) (*Comment, error) {
	c := &Comment{}
	err := data.CommentGet(c, key)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// CommentGetTree retrieves  a single comment and it's child comments
func CommentGetTree(key data.UUID, limit int, sort string) (*CommentTree, error) {
	c := &CommentTree{}
	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(commentMaxRetrieve)))

	err := data.CommentGetTree(c, key, limit+1, sort)
	if err != nil {
		return nil, err
	}
	c.setMoreChildren(limit)

	return c, nil
}

// CommentsGet retrieves comments from a given post or parent comment
func CommentsGet(post *Post, parent *Comment, from, limit int, sort string) (comments []CommentTree, more bool, err error) {

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(commentMaxRetrieve)))
	from = int(math.Max(0, float64(from)))

	parentKey := post.Key
	if parent != nil {
		parentKey = parent.Key
	}

	err = data.CommentsGet(&comments, post.Key, parentKey, from, limit+1, sort)
	if err == data.ErrNotFound {
		return comments, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	for i := range comments {
		comments[i].setMoreChildren(limit)
	}

	more = len(comments) > limit
	if more {
		comments = comments[:len(comments)-1]
	}

	return comments, more, nil
}

// CommentNew adds a new root comment on a post
func CommentNew(who *User, post *Post, comment string) (*Comment, error) {
	if who == nil {
		return nil, ErrCommentNoUser
	}

	c := &Comment{
		PostKey:  post.Key,
		Parent:   post.Key,
		Username: who.Username,
		Comment:  comment,
		post:     post,
		user:     who,
	}

	err := c.validate()
	if err != nil {
		return nil, err
	}

	err = c.insert()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Reply relies to a specific comment
func (c *Comment) Reply(who *User, comment string) (*Comment, error) {
	if who == nil {
		return nil, ErrCommentNoUser
	}

	reply := &Comment{
		PostKey:  c.PostKey,
		Parent:   c.Key,
		Username: who.Username,
		Comment:  comment,
		user:     who,
	}

	err := reply.validate()
	if err != nil {
		return nil, err
	}

	err = reply.insert()
	if err != nil {
		return nil, err
	}

	//notify parent comment of reply
	err = reply.notifyReply()
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (c *Comment) notifyReply() error {
	_, err := c.Post()
	if err != nil {
		return err
	}

	parent, err := c.ParentComment()
	if err == data.ErrNotFound {
		// no parent found, can't notify of reply
		return nil
	}

	if err != nil {
		return err
	}

	if parent.Username == c.Username {
		// don't notify self of reply
		return nil
	}

	u, err := parent.User()
	if err != nil {
		return err
	}

	msgData := struct {
		To      *User
		Post    *Post
		Comment *Comment
	}{
		To:      u,
		Post:    c.post,
		Comment: c,
	}

	sub, msg, err := messages.use("msgCommentReply").Execute(msgData)
	if err != nil {
		return err
	}

	err = notificationNew(data.EmptyKey, parent.Username, sub, msg)
	if err != nil {
		return err
	}

	if u.EmailCommentReply {
		sub, msg, err = messages.use("emailCommentReply").Execute(msgData)
		if err != nil {
			return err
		}
		return u.sendEmail(sub, msg)
	}
	return nil

}

func (c *Comment) validate() error {
	if c.Comment == "" {
		return ErrCommentEmpty
	}
	if len(c.Comment) > commentMaxLength {
		return ErrCommentTooLong
	}

	if c.Parent == data.EmptyUUID {
		// should never happen
		// Parent on root comments is the postkey
		return fail.New("No parent specified for this comment")
	}

	if c.PostKey == data.EmptyUUID {
		return fail.New("No post specified for this comment")
	}

	_, err := c.Post()
	if err != nil {
		return err
	}

	if !c.post.AllowComments {
		return ErrCommentNotAllowed
	}

	if c.post.Status != PostStatusPublished {
		return ErrCommentPostNotPublished
	}

	return nil
}

func (c *Comment) insert() error {
	c.Rev()
	key, err := data.CommentInsert(c)
	if err != nil {
		return err
	}
	c.Key = key

	err = c.sendNotifications()
	if err != nil {
		return err
	}
	return c.sendMentions()
}

// Post retrieves the root post for a given comment
func (c *Comment) Post() (*Post, error) {
	if c.post != nil {
		return c.post, nil
	}

	post, err := PostGet(c.PostKey)
	if err != nil {
		return nil, err
	}
	c.post = post
	return post, nil
}

// ParentComment retrieves the comments parent
func (c *Comment) ParentComment() (*Comment, error) {
	if c.parent != nil {
		return c.parent, nil
	}

	if c.Parent == data.EmptyUUID {
		return nil, data.ErrNotFound
	}

	parent, err := CommentGet(c.Parent)
	if err != nil {
		return nil, err
	}
	c.parent = parent

	return parent, nil
}

func (c *Comment) sendMentions() error {
	users := rxMention.FindAllString(c.Comment, -1)
	for i := range users {
		username := data.NewKey(users[i][1:]) // drop leading @
		if username == c.Username {
			//don't notify self of mention
			continue
		}
		u, err := UserGet(username)
		if err == ErrUserNotFound {
			continue
		}
		if err != nil {
			return err
		}
		err = u.mentionComment(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Comment) sendNotifications() error {
	_, err := c.Post()
	if err != nil {
		return err
	}

	if !c.post.NotifyOnComment {
		return nil
	}

	if c.Username == c.post.Creator {
		//don't notify on your own comments
		return nil
	}

	user, err := c.post.creator()
	if err != nil {
		return err
	}

	msgData := struct {
		To      *User
		Post    *Post
		Comment *Comment
	}{
		To:      user,
		Post:    c.post,
		Comment: c,
	}

	sub, msg, err := messages.use("msgCommentPost").Execute(msgData)
	if err != nil {
		return err
	}

	err = notificationNew(data.EmptyKey, user.Username, sub, msg)
	if err != nil {
		return err
	}

	if user.EmailPostComment {
		sub, msg, err := messages.use("emailCommentPost").Execute(msgData)
		if err != nil {
			return err
		}

		return user.sendEmail(sub, msg)
	}

	return nil
}

// User returns the user on a given comment
func (c *Comment) User() (*User, error) {
	if c.user != nil {
		return c.user, nil
	}

	return UserGet(c.Username)
}
