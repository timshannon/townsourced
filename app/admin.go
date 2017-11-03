// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package app

import (
	"time"

	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

// ErrNotAdmin is when a user fails a login attempt
var ErrNotAdmin = fail.New("You do not have access.")

// AdminStats holds the stats presented on the admin page
type AdminStats struct {
	UserCount      int `json:"userCount"`
	UserCountTrend []struct {
		Count int       `json:"count" gorethink:"reduction"`
		Date  time.Time `json:"date" gorethink:"group"`
	} `json:"userCountTrend"`
	UserLast []*User `json:"userLast"`

	TownCount      int `json:"townCount"`
	TownCountTrend []struct {
		Count int       `json:"count" gorethink:"reduction"`
		Date  time.Time `json:"date" gorethink:"group"`
	} `json:"townCountTrend"`
	TownLast []*Town `json:"townLast"`

	PostCount      int `json:"postCount"`
	PostCountTrend []struct {
		Count int       `json:"count" gorethink:"reduction"`
		Date  time.Time `json:"date" gorethink:"group"`
	} `json:"postCountTrend"`
	PostLast []*Post `json:"postLast"`
}

//AdminStatsGet retrieves current admin page stats
// Any errors will be returned as failures as it is assumed
// the user is an admin and can view full errors
func AdminStatsGet(who *User, since time.Time) (*AdminStats, error) {
	stats := &AdminStats{}
	var err error

	if since.IsZero() {
		since = time.Now().AddDate(0, 0, -180)
	}

	if !who.Admin {
		return nil, ErrNotAdmin
	}

	//TODO: gather these on separate goroutines

	//user stats
	stats.UserCount, err = data.UserAllCount()
	if err != nil {
		return nil, fail.NewFromErr(err)
	}

	err = data.AdminLastUsers(&stats.UserLast)
	if err != nil {
		return nil, fail.NewFromErr(err)
	}

	err = data.AdminUserCountTrend(&stats.UserCountTrend, since)
	if err != nil {
		return nil, fail.NewFromErr(err)
	}

	//town stats
	stats.TownCount, err = data.TownAllCount()
	if err != nil {
		return nil, fail.NewFromErr(err)
	}
	err = data.AdminLastTowns(&stats.TownLast)
	if err != nil {
		return nil, fail.NewFromErr(err)
	}
	err = data.AdminTownCountTrend(&stats.TownCountTrend, since)
	if err != nil {
		return nil, fail.NewFromErr(err)
	}

	//post stats
	stats.PostCount, err = data.PostAllCount()
	if err != nil {
		return nil, fail.NewFromErr(err)
	}
	err = data.AdminLastPosts(&stats.PostLast)
	if err != nil {
		return nil, fail.NewFromErr(err)
	}
	err = data.AdminPostCountTrend(&stats.PostCountTrend, since)
	if err != nil {
		return nil, fail.NewFromErr(err)
	}

	return stats, nil
}
