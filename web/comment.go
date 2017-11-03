// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"strconv"
	"time"

	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

type commentInput struct {
	Comment *string `json:"comment,omitempty"`
}

// rate limit the number of new comments that can be created by the same user
var commentNewRequestType = app.RequestType{
	Type:         "commentNew",
	FreeAttempts: 30,
	Scale:        5 * time.Second,
	Range:        5 * time.Minute,
	MaxWait:      1 * time.Minute,
}

const commentListSizeDefault = 40

func commentGet(w http.ResponseWriter, r *http.Request, c context) {
	// ?from=<from>&limit=100
	// ?sort=<new|old>

	values := r.URL.Query()

	limit, err := strconv.Atoi(values.Get("limit"))
	if err != nil {
		limit = commentListSizeDefault
	}

	from, err := strconv.Atoi(values.Get("from"))
	if err != nil {
		from = 0
	}

	post, err := app.PostGet(data.ToUUID(c.params.ByName("post")))
	if err == app.ErrPostNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	var parent *app.Comment
	parentKey := c.params.ByName("comment")
	if parentKey != "" {
		parent, err = app.CommentGet(data.ToUUID(parentKey))
		if errHandled(err, w, r, c) {
			return
		}
	}

	comments, more, err := app.CommentsGet(post, parent, from, limit, values.Get("sort"))
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   comments,
		More:   more,
	})
}

func commentPost(w http.ResponseWriter, r *http.Request, c context) {
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

	input := &commentInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Comment == nil {
		errHandled(fail.New("You must specify some text when commenting", input), w, r, c)
		return
	}

	post, err := app.PostGet(data.ToUUID(c.params.ByName("post")))
	if err == app.ErrPostNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	//Rate limit new comments
	if errHandled(app.AttemptRequest(string(c.session.UserKey), commentNewRequestType), w, r, c) {
		return
	}

	parentKey := c.params.ByName("comment")
	if parentKey != "" {
		comment, err := app.CommentGet(data.ToUUID(parentKey))
		if errHandled(err, w, r, c) {
			return
		}
		newComment, err := comment.Reply(u, *input.Comment)
		if errHandled(err, w, r, c) {
			return
		}

		respondJsendCode(w, &JSend{
			Status: statusSuccess,
			Data:   newComment,
		}, http.StatusCreated)
		return
	}

	comment, err := app.CommentNew(u, post, *input.Comment)
	if errHandled(err, w, r, c) {
		return

	}

	respondJsendCode(w, &JSend{
		Status: statusSuccess,
		Data:   comment,
	}, http.StatusCreated)
}
