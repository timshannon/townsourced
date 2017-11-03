// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/fail"
)

const (
	cookieName = "townsourced"
)

// rate limit login attempts
var userLogonRequest = app.RequestType{
	Type:         "login",
	FreeAttempts: 10,
	Scale:        5 * time.Second,
	Range:        5 * time.Minute,
	MaxWait:      1 * time.Minute,
}

// get a session from the request
func session(r *http.Request) (*app.Session, error) {
	// must iter through all cookies because you can have
	// multiple cookies with the same name
	// the cookie is valid only if the name matches AND it has a value
	cookies := r.Cookies()
	cValue := ""

	for i := range cookies {
		if cookies[i].Name == cookieName {
			if cookies[i].Value != "" {
				cValue = cookies[i].Value
			}
		}
	}

	if cValue == "" {
		return nil, nil
	}

	s, err := app.SessionGet(cValue)
	if err == app.ErrSessionInvalid {
		return nil, nil
	}
	return s, err
}

func handleCSRF(w http.ResponseWriter, r *http.Request, s *app.Session) error {
	if s == nil {
		return nil
	}

	if r.Method != "GET" {
		reqToken := r.Header.Get("X-CSRFToken")
		if reqToken != s.CSRFToken && s.Valid {
			return fail.New("Invalid CSRFToken.  Your session may be invalid.  Try logging in again.", reqToken)
		}
		return nil
		//TODO: Consider resetting CSRF token regularily?
		// On each update? Once a day?
	}

	//Get requests, put CSRF token in header
	w.Header().Add("X-CSRFToken", s.CSRFToken)

	return nil
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, u *app.User, rememberMe bool) error {
	expires := time.Time{}

	if rememberMe {
		expires = time.Now().AddDate(0, 0, 15)
	}

	s, err := app.SessionNew(u, expires, ipAddress(r), r.UserAgent())
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    s.Key,
		HttpOnly: true,
		Path:     "/",
		Secure:   isSSL,
		Expires:  expires,
	}

	http.SetCookie(w, cookie)
	return nil
}

func expireSessionCookie(w http.ResponseWriter, r *http.Request, s *app.Session) {
	cookie, err := r.Cookie(cookieName)

	if err != http.ErrNoCookie {
		if cookie.Value == s.Key {
			cookie := &http.Cookie{
				Name:     cookieName,
				Value:    "",
				HttpOnly: true,
				Path:     "/",
				Secure:   isSSL,
				MaxAge:   0,
			}

			http.SetCookie(w, cookie)
		}
	}
}

type sessionInput struct {
	userInput
	RememberMe bool `json:"rememberMe,omitempty"`
}

// login
func sessionPost(w http.ResponseWriter, r *http.Request, c context) {
	if c.session != nil {
		//If previous session still exists, log out so it can't be used again
		go func(session *app.Session) {
			err := session.Logout()
			if err != nil {
				log.WithField("session", session).Error("Error logging out session when trying to log into a new session")
			}
		}(c.session)
	}

	input := &sessionInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Email == nil && input.Username == nil {
		errHandled(fail.New("An email address or username is required", input), w, r, c)
		return
	}

	usernameOrEmail := ""
	if input.Email != nil {
		usernameOrEmail = *input.Email
	}

	if input.Username != nil {
		usernameOrEmail = string(*input.Username)
	}

	if input.Password == nil {
		errHandled(fail.New("You must specify a password", input), w, r, c)
		return
	}

	// rate limit login requests
	if errHandled(app.AttemptRequest(ipAddress(r), userLogonRequest), w, r, c) {
		return
	}

	u, err := app.UserLogin(usernameOrEmail, *input.Password)
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(setSessionCookie(w, r, u, input.RememberMe), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

// logout
func sessionDelete(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	expireSessionCookie(w, r, c.session)

	if errHandled(c.session.Logout(), w, r, c) {
		return
	}
	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}
