// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"time"

	rt "git.townsourced.com/townsourced/gorethink"
)

// A tempToken is a temporary token set to expire
// that can have data attached to it.
// Usually used for one off CSRF token request or password resets

func init() {
	tables = append(tables, tblTempToken)
}

var tblTempToken = &table{
	name: "tempToken",
}

type cacheTempToken struct {
	token   string
	expires time.Duration
}

func (t *cacheTempToken) key() string {
	return t.token
}

func (t *cacheTempToken) source(result interface{}) error {
	c, err := tblTempToken.Get(t.token).Default([]interface{}{}).Field("Data").Run(session)
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

	err = c.One(result)
	if err == rt.ErrEmptyResult {
		return ErrNotFound
	}
	return err
}

func (t *cacheTempToken) expiration() time.Duration {
	return t.expires
}

func (t *cacheTempToken) refresh() {}
func (t *cacheTempToken) dependents() []cacher {
	return nil
}

// TempTokenGet gets a tempory token
func TempTokenGet(tokenData interface{}, token string) error {
	return cacheGet(&cacheTempToken{
		token: token,
	}, tokenData)
}

// TempTokenSet sets a session in the memcache if expires is less than 30 minutes, otherwise it goes in rethink
func TempTokenSet(tokenData interface{}, token string, expires time.Duration) error {
	if expires > 30*time.Minute {
		return wErr(tblTempToken.Insert(struct {
			Key     string
			Expires time.Duration
			Data    interface{}
		}{
			Key:     token,
			Expires: expires,
			Data:    tokenData,
		}).RunWrite(session))
	}
	return cacheSet(&cacheTempToken{
		token:   token,
		expires: expires,
	}, tokenData)
}

// TempTokenExpire will expire the temp token with the passed in ID
func TempTokenExpire(token string) error {
	err := TempTokenSet(nil, token, 0)
	if err != nil {
		return err
	}

	err = wErr(tblTempToken.Get(token).Delete(rt.DeleteOpts{Durability: "soft"}).RunWrite(session))
	if err != nil && err != ErrNotFound {
		return err
	}
	return nil
}
