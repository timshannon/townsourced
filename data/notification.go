// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"time"

	rt "git.townsourced.com/townsourced/gorethink"
)

func init() {
	tables = append(tables, tblNotification)
}

var tblNotification = &table{
	name: "notification",
	indexes: []index{
		index{
			name: "Username_When",
			indexFunc: func(row rt.Term) interface{} {
				return []interface{}{row.Field("Username"), row.Field("When")}
			},
		},
		index{
			name: "From_When",
			indexFunc: func(row rt.Term) interface{} {
				return []interface{}{row.Field("From"), row.Field("When")}
			},
		},
	},
}

// NotificationInsert inserts a new user notification into the database
func NotificationInsert(notification interface{}) error {
	return wErr(tblNotification.Insert(notification).RunWrite(session))
}

// NotificationGetUnread retrieves all unread notifications for a user
func NotificationGetUnread(result interface{}, username Key, since time.Time, limit int) error {
	return notificationsGet(result, username, since, limit, true)
}

func notificationsGet(result interface{}, username Key, since time.Time, limit int, unread bool) (err error) {
	var sinceOp interface{} = rt.MaxVal

	if !since.IsZero() {
		sinceOp = since
	}

	trm := tblNotification.Between([]interface{}{username, rt.MinVal}, []interface{}{username, sinceOp},
		rt.BetweenOpts{
			Index:     "Username_When",
			LeftBound: "open",
		}).OrderBy(rt.OrderByOpts{
		Index: rt.Desc("Username_When"),
	})
	if unread {
		trm = trm.Filter(map[string]interface{}{
			"Read": false,
		})

	}
	c, err := trm.Limit(limit).Run(session)

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

//NotificationGetAll retrieves all notifications for a user
func NotificationGetAll(result interface{}, username Key, since time.Time, limit int) error {
	return notificationsGet(result, username, since, limit, false)
}

// NotificationUnreadCount retrieves the number of unread notifications for a user
func NotificationUnreadCount(username Key) (int, error) {
	var count int

	c, err := tblNotification.Between([]interface{}{username, rt.MinVal},
		[]interface{}{username, rt.MaxVal},
		rt.BetweenOpts{
			Index: "Username_When",
		}).Filter(map[string]interface{}{
		"Read": false,
	}).Count().Run(session)

	if err != nil {
		return -1, err
	}

	err = c.One(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

// NotificationGet retrieves a specific notification for a user
func NotificationGet(result interface{}, notificationKey UUID) error {
	c, err := tblNotification.Get(notificationKey).Run(session)

	if err != nil {
		return err
	}

	if c.IsNil() {
		return ErrNotFound
	}

	return c.One(result)
}

// NotificationUpdate updates a single notification
func NotificationUpdate(notification interface{}, key UUID) error {
	return wErr(tblNotification.Get(key).Update(notification).RunWrite(session))
}

// NotificationUpdateUnread updates all unread notifications
func NotificationUpdateUnread(notification interface{}, username Key) error {
	return wErr(tblNotification.Between([]interface{}{username, rt.MinVal},
		[]interface{}{username, rt.MaxVal},
		rt.BetweenOpts{
			Index: "Username_When",
		}).Filter(map[string]interface{}{
		"Read": false,
	}).Update(notification).RunWrite(session))
}

// NotificationsGetSent gets all sent notifications for a user
func NotificationsGetSent(result interface{}, username Key, since time.Time, limit int) (err error) {
	var sinceOp interface{} = rt.MaxVal

	if !since.IsZero() {
		sinceOp = since
	}

	trm := tblNotification.Between([]interface{}{username, rt.MinVal}, []interface{}{username, sinceOp},
		rt.BetweenOpts{
			Index:     "From_When",
			LeftBound: "open",
		}).OrderBy(rt.OrderByOpts{
		Index: rt.Desc("From_When"),
	})

	c, err := trm.Limit(limit).Run(session)

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
