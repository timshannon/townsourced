// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"

	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/fail"
)

func googleGet(w http.ResponseWriter, r *http.Request, c context) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	code := r.URL.Query().Get("code")

	if code == "" || redirectURI == "" {
		clientID, authURL, err := app.GoogleOpenIDConfig()
		if errHandled(err, w, r, c) {
			return
		}

		respondJsend(w, &JSend{
			Status: statusSuccess,
			Data: map[string]string{
				"clientID": clientID,
				"authURL":  authURL,
			},
		})
		return
	}
	if c.session == nil {
		u, err := app.GoogleUser(code, redirectURI)
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

	err = u.LinkGoogle(redirectURI, code)
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

type googleInput struct {
	UserID      string `json:"userID"`
	IDToken     string `json:"idToken"`
	AccessToken string `json:"accessToken"`
	Username    string `json:"username"`
	Email       string `json:"email"`
}

func googlePost(w http.ResponseWriter, r *http.Request, c context) {
	if demoMode {
		if errHandled(fail.New("Signups are currently disabled.  Contact info@townsourced.com "+
			"for more information"), w, r, c) {
			return
		}

	}

	input := &googleInput{}

	if errHandled(parseInput(r, input), w, r, c) {
		return
	}

	u, err := app.GoogleNewUser(input.Username, input.Email, input.UserID, input.IDToken, input.AccessToken)
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

func googleDelete(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	err = u.DisconnectGoogle()
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
