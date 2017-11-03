// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"

	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/fail"
)

type twitterInput struct {
	Token    *string `json:"token"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
}

func twitterGet(w http.ResponseWriter, r *http.Request, c context) {
	code := r.URL.Query().Get("code")
	token := r.URL.Query().Get("token")
	if token == "" {
		errHandled(fail.New("token is required"), w, r, c)
		return
	}
	if code == "" {
		errHandled(fail.New("code is required"), w, r, c)
		return
	}

	if c.session == nil {
		u, err := app.TwitterGetUser(token, code)
		if errHandled(err, w, r, c) {
			return
		}

		if errHandled(setSessionCookie(w, r, u, true), w, r, c) {
			return
		}

		respondJsend(w, &JSend{
			Status: statusSuccess,
		})
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	err = u.LinkTwitter(token, code)
	if errHandled(err, w, r, c) {
		return
	}

	err = u.Update()
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})

}

func twitterPost(w http.ResponseWriter, r *http.Request, c context) {
	if demoMode {
		if errHandled(fail.New("Signups are currently disabled.  Contact info@townsourced.com "+
			"for more information"), w, r, c) {
			return
		}

	}

	input := &twitterInput{}

	if errHandled(parseInput(r, input), w, r, c) {
		return
	}

	if input.Token == nil {
		errHandled(fail.New("token is a required input", input), w, r, c)
		return
	}

	if input.Username == nil && input.Email == nil {
		uri, err := app.TwitterGetLoginURL(siteURL(r, "3rdparty/").String(), *input.Token)
		if errHandled(err, w, r, c) {
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
			Data:   uri,
		})
		return
	}

	username := ""
	email := ""

	if input.Username != nil {
		username = *input.Username
	}

	if input.Email != nil {
		email = *input.Email
	}

	u, err := app.TwitterNewUser(username, email, *input.Token)
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(setSessionCookie(w, r, u, true), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})

}

func twitterDelete(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	err = u.DisconnectTwitter()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(u.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}
