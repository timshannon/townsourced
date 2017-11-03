// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"time"

	"git.townsourced.com/townsourced/httprouter"
	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

var userPassworkTokenRate = app.RequestType{
	Type:         "userPasswordToken",
	FreeAttempts: 10,
	Scale:        5 * time.Second,
	Range:        15 * time.Minute,
	MaxWait:      1 * time.Minute,
}

func forgotPasswordTemplate(w http.ResponseWriter, r *http.Request, c context) {
	token := c.params.ByName("token")

	var user *app.User

	if token != "" {
		if errHandledPage(app.AttemptRequest(ipAddress(r), userPassworkTokenRate), w, r, c) {
			return
		}

		username, err := app.RetrievePasswordToken(token)
		if err != data.ErrNotFound && errHandledPage(err, w, r, c) {
			return
		}

		if username != data.EmptyKey {
			user, err = app.UserGet(username)
			if err != data.ErrNotFound && errHandledPage(err, w, r, c) {
				return
			}
			user.ClearPrivate()
		}
	}

	err := w.(*templateWriter).execute("FORGOTPASSWORD", struct {
		User  *app.User
		Token string
	}{
		User:  user,
		Token: token,
	})
	if err != nil {
		log.Errorf("Error executing FORGOTPASSWORD template: %s", err)
	}
}

func resetPassword(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//faked context to prevent csrf issues
	c := context{
		params: p,
	}

	input := &userInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	token := c.params.ByName("token")

	if input.NewPassword == nil || token == "" {
		errHandled(fail.New("newPassword and password token are required to reset a password", input), w, r, c)
		return
	}

	if errHandled(app.AttemptRequest(ipAddress(r), userPassworkTokenRate), w, r, c) {
		return
	}

	u, err := app.ResetPassword(token, *input.NewPassword)

	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(setSessionCookie(w, r, u, false), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func forgotPassword(w http.ResponseWriter, r *http.Request, c context) {
	input := &userInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	usernameOrEmail := ""
	if input.Email != nil {
		usernameOrEmail = *input.Email
	}

	if input.Username != nil {
		usernameOrEmail = string(*input.Username)
	}

	if errHandled(app.ForgotPassword(usernameOrEmail), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func confirmEmailTemplate(w http.ResponseWriter, r *http.Request, c context) {
	token := c.params.ByName("token")

	if token == "" {
		four04(w, r)
		return
	}

	success := true

	err := app.ConfirmEmail(token)
	if err != nil {
		if !fail.IsFail(err) {
			errHandledPage(err, w, r, c)
			return
		}
		success = false
	}

	err = w.(*templateWriter).execute("CONFIRMEMAIL", success)
	if err != nil {
		log.Errorf("Error executing CONFIRMEMAIL template: %s", err)
	}
}

func userConfirmEmail(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(u.SendEmailConfirmation(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}
