// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"path"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/data/private"
	"github.com/timshannon/townsourced/fail"
)

type postInput struct {
	Title            *string     `json:"title,omitempty"`
	Content          *string     `json:"content,omitempty"`
	Category         *string     `json:"category,omitempty"`
	Format           *string     `json:"format,omitempty"`
	TownKeys         []data.Key  `json:"townKeys"`
	Images           []data.UUID `json:"images"`
	FeaturedImage    *data.UUID  `json:"featuredImage,omitempty"`
	VerTag           *string     `json:"vertag,omitempty"`
	Draft            bool        `json:"draft,omitempty"`
	Close            bool        `json:"close,omitempty"`
	ReOpen           bool        `json:"reopen,omitempty"`
	Unpublish        bool        `json:"unpublish,omitempty"`
	ModeratorTown    *data.Key   `json:"moderatorTown,omitempty"`
	ModeratorReason  *string     `json:"moderatorReason,omitempty"`
	RemoveModeration bool        `json:"removeModeration,omitempty"`
	NotifyOnComment  *bool       `json:"notifyOnComment,omitempty"`
	AllowComments    *bool       `json:"allowComments,omitempty"`
	Report           *string     `json:"report,omitempty"`
}

// rate limit the number of new posts that can be created by the same user
var postNewRequestType = app.RequestType{
	Type:         "postNew",
	FreeAttempts: 10,
	Scale:        5 * time.Second,
	Range:        10 * time.Minute,
	MaxWait:      1 * time.Minute,
}

func postTemplate(w http.ResponseWriter, r *http.Request, c context) {
	key := c.params.ByName("post")
	post, err := app.PostGet(data.ToUUID(key))
	if err == app.ErrPostNotFound {
		four04(w, r)
		return
	}
	if errHandledPage(err, w, r, c) {
		return
	}

	if post.Status == app.PostStatusDraft {
		//Redirect to edit post
		http.Redirect(w, r, path.Join("..", "editpost", key), http.StatusTemporaryRedirect)
		return
	}

	var user *app.User
	if c.session != nil {
		user, err = c.session.User()
		if errHandledPage(err, w, r, c) {
			return
		}
	}

	visible, err := post.Visible(user)
	if errHandledPage(err, w, r, c) {
		return
	}
	if !visible {
		four04(w, r)
		return
	}

	var comments []app.CommentTree
	var commentTree *app.CommentTree
	var values = r.URL.Query()

	more := false

	commentContext := c.params.ByName("comment")

	if commentContext == "" {
		comments, more, err = app.CommentsGet(post, nil, 0, commentListSizeDefault, values.Get("sort"))
		if errHandledPage(err, w, r, c) {
			return
		}
	} else {
		commentTree, err = app.CommentGetTree(data.ToUUID(commentContext), commentListSizeDefault,
			values.Get("sort"))
		if err == data.ErrNotFound {
			four04(w, r)
			return
		}
		if errHandledPage(err, w, r, c) {
			return
		}
	}

	description := post.RawContent()
	canonicalURL := siteURL(r, path.Join("post", key)).String()
	openGraph := map[string]string{
		"fb:app_id":      private.FacebookAppID,
		"og:url":         canonicalURL,
		"og:description": description,
	}

	schema := map[string]string{
		"name":        post.Title,
		"description": description,
	}

	if post.FeaturedImage != data.EmptyUUID {
		openGraph["og:image"] = siteURL(r, path.Join("api/v1/image", data.FromUUID(post.FeaturedImage))).String()
		openGraph["og:image:url"] = siteURL(r, path.Join("api/v1/image", data.FromUUID(post.FeaturedImage))).String()
		schema["image"] = siteURL(r, path.Join("api/v1/image", data.FromUUID(post.FeaturedImage))).String()
	} else {
		openGraph["og:image"] = siteURL(r, "/images/shared_placeholder.png").String()
		openGraph["og:image:url"] = siteURL(r, "/images/shared_placeholder.png").String()
		schema["image"] = siteURL(r, "/images/shared_placeholder.png").String()
	}

	err = w.(*templateWriter).execute("POST",
		struct {
			Post           *app.Post
			User           *app.User
			Comments       []app.CommentTree
			CommentContext *app.CommentTree
			Description    string
			More           bool
			OpenGraph      map[string]string
			Schema         map[string]string
			SchemaType     string
			URL            string
		}{
			Post:           post,
			User:           user,
			Comments:       comments,
			CommentContext: commentTree,
			Description:    description,
			More:           more,
			OpenGraph:      openGraph,
			Schema:         schema,
			SchemaType:     postToSchemaType(post),
			URL:            canonicalURL,
		})
	if err != nil {
		log.Errorf("Error executing post template: %s", err)
	}
}

func editPostTemplate(w http.ResponseWriter, r *http.Request, c context) {
	key := c.params.ByName("post")

	var post *app.Post

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	var user *app.User
	var err error
	if c.session != nil {
		user, err = c.session.User()
		if errHandledPage(err, w, r, c) {
			return
		}
	}

	if key == "" {
		//New post
		err := w.(*templateWriter).execute("EDITPOST", struct {
			Post       *app.Post
			User       *app.User
			ShareError bool
		}{
			Post: post,
			User: user,
		})
		if err != nil {
			log.Errorf("Error executing editpost template: %s", err)
		}
		return
	}

	post, err = app.PostGet(data.ToUUID(key))
	if err == app.ErrPostNotFound {
		four04(w, r)
		return
	}
	if errHandledPage(err, w, r, c) {
		return
	}

	visible, err := post.Visible(user)
	if errHandled(err, w, r, c) {
		return
	}
	if !visible {
		four04(w, r)
		return
	}

	err = post.CanEdit(user)
	if err == app.ErrPostNotDraft {
		http.Redirect(w, r, path.Join("..", "post", key), http.StatusTemporaryRedirect)
		return
	}

	if err == app.ErrPostNotOwner {
		four04(w, r)
		return
	}

	err = w.(*templateWriter).execute("EDITPOST", struct {
		Post       *app.Post
		User       *app.User
		ShareError bool
	}{
		Post: post,
		User: user,
	})

	if err != nil {
		log.Errorf("Error executing editpost template: %s", err)
	}
}

func postGet(w http.ResponseWriter, r *http.Request, c context) {
	key := c.params.ByName("post")
	post, err := app.PostGet(data.ToUUID(key))
	if err == app.ErrPostNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	var user *app.User
	if c.session != nil {
		user, err = c.session.User()
		if errHandledPage(err, w, r, c) {
			return
		}
	}

	visible, err := post.Visible(user)
	if errHandled(err, w, r, c) {
		return
	}
	if !visible {
		four04(w, r)
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   post,
	})
}

func postPost(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}
	u, err := c.session.User()

	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	input := &postInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Title == nil {
		errHandled(fail.New("A Title is required for a new post", input), w, r, c)
		return
	}

	if input.Content == nil {
		errHandled(fail.New("Content is required for a new post", input), w, r, c)
		return
	}

	format := app.PostFormatStandard

	if input.Format != nil {
		format = *input.Format
	}

	category := ""
	if input.Category != nil {
		category = *input.Category
	}

	featuredImage := data.EmptyUUID
	if input.FeaturedImage != nil {
		featuredImage = *input.FeaturedImage
	}

	notifyOnComment := false
	if input.NotifyOnComment != nil {
		notifyOnComment = *input.NotifyOnComment
	}

	allowComments := false
	if input.AllowComments != nil {
		allowComments = *input.AllowComments
	}

	//Rate limit new posts
	if errHandled(app.AttemptRequest(string(c.session.UserKey), postNewRequestType), w, r, c) {
		return
	}

	post, err := app.PostNew(*input.Title, *input.Content, category, format, u, input.TownKeys, input.Images,
		featuredImage, allowComments, notifyOnComment, input.Draft)
	if errHandled(err, w, r, c) {
		return
	}

	respondJsendCode(w, &JSend{
		Status: statusSuccess,
		Data:   post,
	}, http.StatusCreated)
}

func postPut(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}
	u, err := c.session.User()

	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	key := data.ToUUID(c.params.ByName("post"))
	post, err := app.PostGet(key)
	if err == app.ErrPostNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	input := &postInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	vertag := ""
	if input.VerTag != nil {
		vertag = *input.VerTag
	}

	post.SetVer(vertag)

	if input.Title != nil {
		if errHandled(post.SetTitle(u, *input.Title), w, r, c) {
			return
		}
	}

	if input.Content != nil {
		if errHandled(post.SetContent(u, *input.Content), w, r, c) {
			return
		}
	}

	if input.Category != nil {
		if errHandled(post.SetCategory(u, *input.Category), w, r, c) {
			return
		}
	}

	if input.Format != nil {
		if errHandled(post.SetFormat(u, *input.Format), w, r, c) {
			return
		}
	}

	if input.TownKeys != nil {
		if errHandled(post.SetTowns(u, input.TownKeys), w, r, c) {
			return
		}
	}

	if input.Images != nil {
		if len(input.Images) > 0 && input.FeaturedImage == nil {
			errHandled(fail.New("You must select a featured image when setting the image list in a post",
				input), w, r, c)
			return
		}

		featImg := data.EmptyUUID
		if input.FeaturedImage != nil {
			featImg = *input.FeaturedImage
		}

		if errHandled(post.SetImages(u, input.Images, featImg), w, r, c) {
			return
		}
	}

	if input.ModeratorReason != nil && input.ModeratorTown != nil {
		town, err := app.TownGet(*input.ModeratorTown)
		if errHandled(err, w, r, c) {
			return
		}

		if errHandled(post.Moderate(town, u, *input.ModeratorReason), w, r, c) {
			return
		}
	}

	if input.RemoveModeration && input.ModeratorTown != nil {
		town, err := app.TownGet(*input.ModeratorTown)
		if errHandled(err, w, r, c) {
			return
		}
		if errHandled(post.RemoveModeration(town, u), w, r, c) {
			return
		}
	}

	if input.Report != nil {
		if errHandled(post.Report(u, *input.Report), w, r, c) {
			return
		}
	}

	if input.AllowComments != nil {
		if errHandled(post.SetAllowComments(u, *input.AllowComments), w, r, c) {
			return
		}
	}

	if input.NotifyOnComment != nil {
		if errHandled(post.SetNotifyOnComment(u, *input.NotifyOnComment), w, r, c) {
			return
		}
	}

	if !input.Draft {
		if input.Close {
			if errHandled(post.Close(u), w, r, c) {
				return
			}
		} else if post.Status == app.PostStatusDraft {
			if errHandled(post.Publish(u), w, r, c) {
				return
			}
		}
	}

	if input.ReOpen {
		if errHandled(post.ReOpen(u), w, r, c) {
			return
		}
	}

	if input.Unpublish {
		if errHandled(post.Unpublish(u), w, r, c) {
			return
		}
	}

	if errHandled(post.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func postToSchemaType(p *app.Post) string {
	switch p.Category {
	case "buysell":
		return "http://schema.org/Offer"
	case "jobs":
		return "https://schema.org/JobPosting"
	case "event":
		return "https://schema.org/Event"
	case "notice":
		return "https://schema.org/Action"
	case "housing":
		return "https://schema.org/Residence"
	case "volunteer":
		return "https://schema.org/Service"
	default:
		return "Thing"

	}
}
