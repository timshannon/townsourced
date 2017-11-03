// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"time"

	log "git.townsourced.com/townsourced/logrus"
)

func init() {
	tables = append(tables, tblSession)
}

var tblSession = &table{
	name: "session",
}

type cacheSession struct {
	sessionKey string
	expires    time.Time
}

func (s *cacheSession) key() string {
	return s.sessionKey
}

func (s *cacheSession) source(result interface{}) error {
	c, err := tblSession.Get(s.key()).Run(session)
	if err != nil {
		return err
	}
	if c.IsNil() {
		return ErrNotFound
	}

	return c.One(result)
}

func (s *cacheSession) expiration() time.Duration {
	if s.expires.IsZero() {
		return 0
	}
	return s.expires.Sub(time.Now())
}

func (s *cacheSession) refresh() {
	err := cacheRefresh(s, map[string]interface{}{})
	if err != nil {
		log.Errorf("Error refreshing session cache. Error: %s", err)
	}
}

func (s *cacheSession) dependents() []cacher {
	return nil
}

// SessionGet retrieves a session
func SessionGet(result interface{}, sessionKey string) error {
	return cacheGet(&cacheSession{
		sessionKey: sessionKey,
	}, result)
}

// SessionInsert inserts a session
func SessionInsert(s interface{}, sessionKey string, expires time.Time) error {
	r, err := tblSession.Insert(s).RunWrite(session)

	err = wErr(r, err)
	if err != nil {
		return err
	}

	go func(s interface{}, sessionKey string, expires time.Time) {
		cerr := cacheSet(&cacheSession{
			sessionKey: sessionKey,
			expires:    expires,
		}, s)
		if cerr != nil {
			log.WithField("SessionData", s).Errorf("Error setting cache for sessionkey %s. Error: %s", sessionKey, cerr)
		}
	}(s, sessionKey, expires)
	return nil
}

// SessionUpdate updates a session
func SessionUpdate(s interface{}, sessionKey string, expires time.Time) error {
	r, err := tblSession.Get(sessionKey).Update(s).RunWrite(session)

	err = wErr(r, err)
	if err != nil {
		return err
	}

	go func(s interface{}, sessionKey string, expires time.Time) {
		cerr := cacheSet(&cacheSession{
			sessionKey: sessionKey,
			expires:    expires,
		}, s)
		if cerr != nil {
			log.WithField("SessionData", s).Errorf("Error updating cache for sessionkey %s. Error: %s", sessionKey, cerr)
		}
	}(s, sessionKey, expires)
	return nil
}
