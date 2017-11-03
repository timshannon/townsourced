// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"errors"
	"strings"
	"time"

	rt "git.townsourced.com/townsourced/gorethink"
)

func init() {
	tables = append(tables, tblComment)
}

// CommentMaxDepth is the max depth at which to retrieve nested comment children
const CommentMaxDepth = 7

/* CommentSort sets the sort order on comments */
const (
	CommentSortNew = "new"
	CommentSortOld = "old"
)

var tblComment = &table{
	name: "comment",
	indexes: []index{
		index{
			name: "Post_Parent",
			indexFunc: func(row rt.Term) interface{} {
				return []interface{}{row.Field("PostKey"), row.Field("Parent")}
			},
		},
		index{
			name: "Username",
			indexFunc: func(row rt.Term) interface{} {
				return []interface{}{row.Field("Username"), row.Field("Updated")}
			},
		},
	},
}

// CommentGet retrieves a single comment
func CommentGet(result interface{}, key UUID) error {
	c, err := tblComment.Get(key).Run(session)
	if err != nil {
		return err
	}

	if c.IsNil() {
		return ErrNotFound
	}

	return c.One(result)
}

// CommentGetTree retrieves a single comment and it's children
func CommentGetTree(result interface{}, key UUID, limit int, sort string) error {
	c, err := commentChildrenTerm(tblComment.Get(key), limit, 0, sort).Run(session)
	if err == rt.ErrEmptyResult {
		return ErrNotFound
	}

	if err != nil {
		return err
	}

	if c.IsNil() {
		return ErrNotFound
	}

	err = c.One(result)
	if err == rt.ErrEmptyResult {
		return ErrNotFound
	}

	return err

}

// CommentsGet retrieves a set of comments
func CommentsGet(result interface{}, postKey, parent UUID, from, limit int, sort string) (err error) {

	trm := tblComment.GetAllByIndex("Post_Parent", []interface{}{postKey, parent}).
		Skip(from).Limit(limit).OrderBy(commentOrderByTerm(sort))

	trm = commentChildrenTerm(trm, limit, 0, sort)
	c, err := trm.Run(session)

	if err != nil {
		return err
	}

	defer func() {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if c.IsNil() {
		return ErrNotFound
	}

	err = c.All(result)
	if err == rt.ErrEmptyResult {
		return ErrNotFound
	}

	return err
}

// recursively builds children merge query
// once max depth is reached, a last merge is done to check for additional children
func commentChildrenTerm(trm rt.Term, limit, depth int, sort string) rt.Term {
	// default to empty list so that merge doesn't run against a null
	trm = trm.Default([]interface{}{})

	if depth == CommentMaxDepth {
		return trm.Merge(func(row rt.Term) interface{} {
			return map[string]interface{}{
				"HasChildren": tblComment.GetAllByIndex("Post_Parent",
					[]interface{}{row.Field("PostKey"), row.Field("Key")}).
					Count().Gt(0),
			}
		})
	}

	depth++
	return trm.Merge(func(row rt.Term) interface{} {
		return map[string]interface{}{
			"Children": commentChildrenTerm(tblComment.GetAllByIndex("Post_Parent",
				[]interface{}{row.Field("PostKey"), row.Field("Key")}).
				Limit(limit).OrderBy(commentOrderByTerm(sort)).CoerceTo("array"), limit, depth, sort),
		}
	})
}

func commentOrderByTerm(sortType string) rt.Term {
	switch strings.ToLower(sortType) {
	case CommentSortNew:
		return rt.Desc("Updated")
	case CommentSortOld:
		return rt.Asc("Updated")
	default:
		return rt.Asc("Updated")
	}
}

// CommentInsert inserts a new user notification into the database
func CommentInsert(comment interface{}) (UUID, error) {
	w, err := tblComment.Insert(comment).RunWrite(session)
	err = wErr(w, err)
	if err != nil {
		return EmptyUUID, err
	}

	if len(w.GeneratedKeys) != 1 {
		return EmptyUUID, errors.New("No new key generated for this comment")
	}

	return UUID(w.GeneratedKeys[0]), nil
}

// CommentUpdate updates an existing comment
func CommentUpdate(comment interface{}, key UUID) error {
	return tryUpdateVersion(tblComment.Get(key), comment)
}

// CommentsGetByUser retrieves a set of comments posted by a given user
func CommentsGetByUser(result interface{}, username Key, public bool, since time.Time, limit int) (err error) {
	var sinceOp interface{} = rt.MaxVal

	if !since.IsZero() {
		sinceOp = since
	}

	trm := tblComment.Between([]interface{}{username, rt.MinVal}, []interface{}{username, sinceOp},
		rt.BetweenOpts{
			Index:     "Username",
			LeftBound: "open",
		}).OrderBy(rt.OrderByOpts{
		Index: rt.Desc("Username"),
	})

	if public {
		trm = trm.Filter(func(comment rt.Term) rt.Term {
			var post = tblPost.Get(comment.Field("PostKey"))
			return postNotModeratedInAllTowns(post).And(postIsPublic(post))
		})
	}

	c, err := trm.Limit(limit).Run(session)

	if err != nil {
		return err
	}

	defer func() {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if c.IsNil() {
		return ErrNotFound
	}

	return c.All(result)
}
