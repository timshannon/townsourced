// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"

	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/fail"
)

func facebookGet(w http.ResponseWriter, r *http.Request, c context) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	code := r.URL.Query().Get("code")

	if code == "" || redirectURI == "" {
		appID := app.FacebookAppID()
		respondJsend(w, &JSend{
			Status: statusSuccess,
			Data: map[string]string{
				"appID": appID,
			},
		})
		return

	}

	if c.session == nil {
		u, err := app.FacebookUser(redirectURI, code)
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

	//active session, link to existing user
	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	err = u.LinkFacebook(redirectURI, code)
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

type facebookInput struct {
	UserID    string `json:"userID"`
	UserToken string `json:"userToken"`
	AppToken  string `json:"appToken"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

func facebookPost(w http.ResponseWriter, r *http.Request, c context) {
	if demoMode {
		if errHandled(fail.New("Signups are currently disabled.  Contact info@townsourced.com "+
			"for more information"), w, r, c) {
			return
		}

	}

	input := &facebookInput{}

	if errHandled(parseInput(r, input), w, r, c) {
		return
	}

	u, err := app.FacebookNewUser(input.Username, input.Email, input.UserID, input.UserToken, input.AppToken)
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

func facebookDelete(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	err = u.DisconnectFacebook()
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
