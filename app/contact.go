// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/timshannon/townsourced/app/email"
	"github.com/timshannon/townsourced/fail"
)

// ContactMessage send a contact message email to the townsourced info inbox
func ContactMessage(fromEmail, toEmail, subject, message string) error {
	if strings.TrimSpace(fromEmail) == "" {
		return fail.New("You must provide an email address")
	}

	if strings.TrimSpace(subject) == "" {
		return fail.New("You must provide a subject")
	}

	if strings.TrimSpace(message) == "" {
		return fail.New("You must provide a message")
	}

	reply := "\n\nOn " + time.Now().Format("Mon, Jan 02, 2006 at 03:04 PM") + `, <` + fromEmail + `> wrote:` +
		"\n\n> " + strings.Replace(message, "\n", "\n> ", -1)

	sub, body, err := messages.use("emailContact").Execute(struct {
		From    string
		Subject string
		Message string
		Reply   string
	}{
		From:    fromEmail,
		Subject: subject,
		Message: message,
		Reply:   reply,
	})
	if err != nil {
		return err
	}

	err = email.Send(email.DefaultFrom, &mail.Address{
		Address: toEmail,
	}, sub, body)

	if err != nil {
		return fmt.Errorf("Error sending email confirmation: %s", err)
	}

	return nil
}
