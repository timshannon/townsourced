// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"math"
	"time"

	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

// Notification is a notification sent to a user.  This can be a message from townsourced
// telling them they've just gotten a comment on their post, or a direct message from another
// user, or anything that needs to be sent to a user
type Notification struct {
	Key      data.UUID `json:"key,omitempty" gorethink:",omitempty"`
	Username data.Key  `json:"username,omitempty" gorethink:",omitempty"`
	From     data.Key  `json:"from,omitempty" gorethink:",omitempty"`
	Subject  string    `json:"subject,omitempty" gorethink:",omitempty"`
	Message  string    `json:"message,omitempty" gorethink:",omitempty"`
	When     time.Time `json:"when,omitempty" gorethink:",omitempty"`
	Read     bool      `json:"read,omitempty"`
}

// ErrNotificationInvalidUser is the error returned when a notification is sent to an invalid user
var ErrNotificationInvalidUser = fail.New("Invalid recipient user for the notification")

const (
	notificationMaxLimit = 500 //max # of notifications that can be requested at a time
)

// new notifications should be generated from the application layer
// so this should not be exported
func notificationNew(from, to data.Key, subject, message string) error {
	if to == data.EmptyKey {
		return ErrNotificationInvalidUser
	}

	notification := &Notification{
		Username: to,
		From:     from,
		Subject:  subject,
		Message:  message,
		When:     time.Now(),
		Read:     false,
	}

	return data.NotificationInsert(notification)
}

// UnreadNotifications returns all unread notifications for a user
func (u *User) UnreadNotifications(since time.Time, limit int) ([]Notification, error) {
	var unread []Notification

	// limit must be between 1 and max
	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(notificationMaxLimit)))

	err := data.NotificationGetUnread(&unread, u.Username, since, limit)
	if err == data.ErrNotFound {
		return unread, nil
	}
	if err != nil {
		return nil, err
	}
	return unread, nil
}

// AllNotifications returns all unread notifications for a user
func (u *User) AllNotifications(since time.Time, limit int) ([]Notification, error) {
	var unread []Notification

	// limit must be between 1 and max
	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(notificationMaxLimit)))

	err := data.NotificationGetAll(&unread, u.Username, since, limit)
	if err == data.ErrNotFound {
		return unread, nil
	}
	if err != nil {
		return nil, err
	}
	return unread, nil
}

// UnreadNotificationCount returns the number of unread notifications for a user
func (u *User) UnreadNotificationCount() (int, error) {
	return data.NotificationUnreadCount(u.Username)
}

// SentNotifications returns all notifications sent by a user
func (u *User) SentNotifications(since time.Time, limit int) ([]Notification, error) {
	var sent []Notification

	// limit must be between 1 and max
	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(notificationMaxLimit)))

	err := data.NotificationsGetSent(&sent, u.Username, since, limit)
	if err == data.ErrNotFound {
		return sent, nil
	}
	if err != nil {
		return nil, err
	}

	// update all notifications are "read"
	for i := range sent {
		sent[i].Read = true
	}
	return sent, nil
}

// Notification will retrieve a specific notification
func (u *User) Notification(key data.UUID) (*Notification, error) {
	n := &Notification{}
	err := data.NotificationGet(n, key)
	if err != nil {
		return nil, err
	}

	if n.Username != u.Username {
		return nil, data.ErrNotFound
	}
	return n, nil
}

// Update updates a specific notification with all the changes that have been applied to it
func (n *Notification) Update() error {
	return data.NotificationUpdate(n, n.Key)
}

// MarkRead marks the notification as read
func (n *Notification) MarkRead() {
	n.Read = true
}

// MarkAllNotificationsAsRead marks all unread notifications as read for a specific user
func (u *User) MarkAllNotificationsAsRead() error {
	n := &Notification{Read: true}

	err := data.NotificationUpdateUnread(n, u.Username)
	if err == data.ErrNotFound {
		// Nothing to update
		return nil
	}
	return err
}
