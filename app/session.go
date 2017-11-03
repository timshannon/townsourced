// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"time"

	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

// Session is an authenticated session into townsourced
type Session struct {
	Key       string //userkey+sessionID - not a traditional data.Key
	UserKey   data.Key
	SessionID string
	CSRFToken string
	Valid     bool
	Expires   time.Time
	When      time.Time
	IPAddress string
	UserAgent string

	user *User
}

// ErrSessionInvalid is returned when a sesssion is invalid or expired
var ErrSessionInvalid = fail.New("Invalid or expired session")

// SessionGet retrieves a session
func SessionGet(sessionKey string) (*Session, error) {
	s := &Session{}
	err := data.SessionGet(s, sessionKey)
	if err == data.ErrNotFound {
		return nil, ErrSessionInvalid
	}
	if err != nil {
		return nil, err
	}
	if !s.Valid || s.Expires.Before(time.Now()) {
		return nil, ErrSessionInvalid
	}

	return s, nil
}

// SessionNew generates a new session for the passed in user
func SessionNew(user *User, expires time.Time, ipAddress, userAgent string) (*Session, error) {
	if expires.IsZero() {
		expires = time.Now().AddDate(0, 0, 3)
	}

	s := &Session{
		UserKey:   user.Username,
		SessionID: Random(128),
		CSRFToken: Random(256),
		Valid:     true,
		Expires:   expires,
		When:      time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	s.Key = string(s.UserKey) + "_" + s.SessionID
	//insert
	err := data.SessionInsert(s, s.Key, s.Expires)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Session) put() error {
	return data.SessionUpdate(s, s.Key, s.Expires)
}

// User Returns the user for the given session, includes private info
func (s *Session) User() (*User, error) {
	if s.user != nil {
		return s.user, nil
	}
	return UserGet(s.UserKey)
}

// Logout logs out of a session
func (s *Session) Logout() error {
	s.Valid = false
	return s.put()
}

// ResetCSRF will generate a new CRSF token, an update the session with it
// allows CSRF token to change more than once per session if need be
func (s *Session) ResetCSRF() error {
	s.CSRFToken = Random(256)
	return s.put()
}
