// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"

	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

type notificationInput struct {
	Key     *data.UUID `json:"key,omitempty"`
	Subject *string    `json:"subject,omitempty"`
	Message *string    `json:"message,omitempty"`
}

func userGetNotifications(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}
	//?count
	//?since=<time>&limit=100&all
	//?sent=true|false
	//?key=<specific notification key>

	if c.session == nil {
		unauthorized(w, r)
		return
	}
	u, err := c.session.User()

	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	values := r.URL.Query()

	key := values.Get("key")

	if key != "" {
		notification, err := u.Notification(data.ToUUID(key))
		if errHandled(err, w, r, c) {
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
			Data:   notification,
		})
		return

	}

	if _, ok := values["count"]; ok {
		count, err := u.UnreadNotificationCount()
		if errHandled(err, w, r, c) {
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
			Data:   count,
		})
		return
	}

	since, limit, err := sinceLimitValues(values, 50)
	if errHandled(err, w, r, c) {
		return
	}

	if _, ok := values["all"]; ok {
		notifications, err := u.AllNotifications(since, limit)
		if errHandled(err, w, r, c) {
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
			Data:   notifications,
		})
		return
	}

	if _, ok := values["sent"]; ok {
		notifications, err := u.SentNotifications(since, limit)
		if errHandled(err, w, r, c) {
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
			Data:   notifications,
		})
		return
	}

	//unread only
	notifications, err := u.UnreadNotifications(since, limit)
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   notifications,
	})
}

func userPutNotifications(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}

	input := &notificationInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()

	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	if input.Key == nil {
		// Mark all as read
		if errHandled(u.MarkAllNotificationsAsRead(), w, r, c) {
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
		})
		return
	}

	n, err := u.Notification(*input.Key)
	if errHandled(err, w, r, c) {
		return
	}

	n.MarkRead()
	if errHandled(n.Update(), w, r, c) {
		return
	}
	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func userPostNotifications(w http.ResponseWriter, r *http.Request, c context) {
	userTo := c.params.ByName("user")

	input := &notificationInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()

	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	if input.Subject == nil || input.Message == nil {
		errHandled(fail.New("Invalid input. Subject, and Message are required fields", input), w, r, c)
		return
	}

	to, err := app.UserGet(data.NewKey(userTo))
	if err == app.ErrUserNotFound {
		errHandled(fail.New("Recipient of message not found", userTo), w, r, c)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(u.SendMessage(to, *input.Subject, *input.Message), w, r, c) {
		return
	}

	respondJsendCode(w, &JSend{
		Status: statusSuccess,
	}, http.StatusCreated)

}
