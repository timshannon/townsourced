// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package data

import (
	"time"

	rt "git.townsourced.com/townsourced/gorethink"
)

// AdminLastUsers returns the last 3 users who signed up
func AdminLastUsers(result interface{}) error {
	c, err := tblUser.OrderBy(rt.OrderByOpts{
		Index: rt.Desc("Created"),
	}).Limit(3).Pluck("Username", "Name").Run(session)

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

	return c.All(result)
}

// AdminLastTowns returns the last 3 towns registered
func AdminLastTowns(result interface{}) error {
	c, err := tblTown.OrderBy(rt.OrderByOpts{
		Index: rt.Desc("Created"),
	}).Limit(3).Pluck("Key", "Name").Run(session)

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

	return c.All(result)
}

// AdminLastPosts returns the last 3 posts published
func AdminLastPosts(result interface{}) error {
	c, err := tblPost.OrderBy(rt.OrderByOpts{
		Index: rt.Desc("Published"),
	}).Limit(3).Pluck("Key", "Title").Run(session)

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

	return c.All(result)
}

// AdminUserCountTrend gets the trend in user counts by day
func AdminUserCountTrend(result interface{}, since time.Time) error {
	c, err := tblUser.Between(since, rt.MaxVal, rt.BetweenOpts{
		Index: "Created",
	}).Group(func(row rt.Term) interface{} {
		return row.Field("Created").Date()
	}).Count().Run(session)

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

	return c.All(result)
}

// AdminTownCountTrend gets the trend in town counts by day
func AdminTownCountTrend(result interface{}, since time.Time) error {
	c, err := tblTown.Between(since, rt.MaxVal, rt.BetweenOpts{
		Index: "Created",
	}).Group(func(row rt.Term) interface{} {
		return row.Field("Created").Date()
	}).Count().Run(session)

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

	return c.All(result)
}

// AdminPostCountTrend gets the trend in post counts by day
func AdminPostCountTrend(result interface{}, since time.Time) error {
	c, err := tblPost.Between(since, rt.MaxVal, rt.BetweenOpts{
		Index: "Published",
	}).Filter(rt.Row.Field("Status").Eq(PostStatusPublished)).Group(func(row rt.Term) interface{} {
		return row.Field("Published").Date()
	}).Count().Run(session)

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

	return c.All(result)
}

//// AdminErrorLog returns the last 100 log entries
//func AdminErrorLog(result interface{}, since time.Time) error {
//c, err := tblLog.Between(since, rt.MaxVal, rt.BetweenOpts{
//Index: "Time",
//}).Limit(100).Run(session)

//if err != nil {
//return err
//}

//defer func() {
//if cerr := c.Close(); cerr != nil && err == nil {
//err = cerr
//}
//}()

//if c.IsNil() {
//return ErrNotFound
//}

//return c.All(result)
//}
