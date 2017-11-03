// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"time"

	"git.townsourced.com/townsourced/townsourced/app"
)

var contactMessageRate = app.RequestType{
	Type:         "contactMessage",
	FreeAttempts: 5,
	Scale:        5 * time.Second,
	Range:        1 * time.Hour,
	MaxWait:      1 * time.Minute,
}

type contactInput struct {
	Email   string `json:"email,omitempty"`
	Subject string `json:"subject,omitempty"`
	Message string `json:"message,omitempty"`
}

const (
	contactEmail = "info@townsourced.com"
)

func contactMessage(w http.ResponseWriter, r *http.Request, c context) {

	input := &contactInput{}
	if errHandled(parseInput(r, input), w, r, c) {
		return
	}

	id := ipAddress(r)
	if c.session != nil {
		id = string(c.session.UserKey)
	}

	if errHandled(app.AttemptRequest(id, contactMessageRate), w, r, c) {
		return
	}

	if errHandled(app.ContactMessage(input.Email, contactEmail, input.Subject, input.Message), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}
