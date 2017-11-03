// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app_test

import (
	"testing"
	"time"

	. "git.townsourced.com/townsourced/check"
	"git.townsourced.com/townsourced/townsourced/app"
)

func (s *AppSuite) TestRateLimit(c *C) {
	reqType := app.RequestType{
		Type:         "TestRateLimit",
		FreeAttempts: 10,
		Scale:        1 * time.Second,
		Range:        2 * time.Minute,
		MaxWait:      10 * time.Second,
	}

	// free attempts
	for i := 0; i < reqType.FreeAttempts; i++ {
		err := app.AttemptRequest("testID", reqType)
		c.Assert(err, Equals, nil)
	}

	if testing.Short() {
		c.Skip("Short testing")
	}

	// delayed attempts
	for i := 0; i < int(reqType.MaxWait/reqType.Scale); i++ {
		err := app.AttemptRequest("testID", reqType)
		c.Assert(err, Equals, nil)
	}

	// errored attempt
	err := app.AttemptRequest("testID", reqType)
	c.Assert(err, ErrorMatches, app.ErrRequestMax.Error())

	// attempt limit should be freed after range expires
	time.Sleep(reqType.Range)
	err = app.AttemptRequest("testID", reqType)
	c.Assert(err, Equals, nil)
}
