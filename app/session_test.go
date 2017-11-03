// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app_test

import (
	"time"

	. "git.townsourced.com/townsourced/check"
	"github.com/timshannon/townsourced/app"
)

const sessionExpireDuration = 30 * time.Second

//Session Test Suite
type SessionSuite struct {
	session *app.Session
	*testData
}

var _ = Suite(&SessionSuite{testData: &testData{}})

func (s *SessionSuite) SetUpSuite(c *C) {
	// add test user
	s.testData.setup(c)
}

func (s *SessionSuite) TearDownSuite(c *C) {
	// delete test user
	s.testData.teardown(c)
}

func (s *SessionSuite) SetUpTest(c *C) {
	// clear memcache before each test
	var err error
	s.session, err = app.SessionNew(s.user, time.Now().Add(sessionExpireDuration), "127.0.0.1", "fake user agent")
	c.Assert(err, Equals, nil)
}

func (s *SessionSuite) TearDownTest(c *C) {
	c.Assert(s.session.Logout(), Equals, nil)
}

func (s *SessionSuite) TestSessionGet(c *C) {
	gotSession, err := app.SessionGet(s.session.Key)
	c.Assert(err, Equals, nil)
	c.Assert(gotSession.Key, Equals, s.session.Key)
	c.Assert(gotSession.UserKey, Equals, s.session.UserKey)
	c.Assert(gotSession.UserAgent, Equals, s.session.UserAgent)
	c.Assert(gotSession.CSRFToken, Equals, s.session.CSRFToken)
}

func (s *SessionSuite) TestSessionUser(c *C) {
	user, err := s.session.User()
	c.Assert(err, Equals, nil)
	c.Assert(user.Username, Equals, s.user.Username)
	c.Assert(user.Email, Equals, s.user.Email)
	c.Assert(user.VerTag, Equals, s.user.VerTag)
}

func (s *SessionSuite) TestSessionLogout(c *C) {
	c.Assert(s.session.Logout(), Equals, nil)
	c.Assert(s.session.Valid, Equals, false)

	//FIXME: Fails tests occasionally, either the cache or the DB isn't getting updated properly
	_, err := app.SessionGet(s.session.Key)
	c.Assert(err, ErrorMatches, app.ErrSessionInvalid.Error())
}

func (s *SessionSuite) TestSessionResetCSRF(c *C) {
	token := s.session.CSRFToken
	c.Assert(s.session.ResetCSRF(), Equals, nil)
	c.Assert(s.session.CSRFToken, Not(Equals), token)
}
