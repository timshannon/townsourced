// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/*Package email is for sending emails for townsourced
There are many different transactional email providers, and I expect which ones we use to change in
the future, so this package will abstract away all access to those providers so nothing in townsourced
should need to change if we change providers
API keys will be stored in the private package with the rest of our keys
*/
package email

import (
	"net/mail"

	log "git.townsourced.com/townsourced/logrus"
	sg "git.townsourced.com/townsourced/sendgrid-go"
	"github.com/timshannon/townsourced/data/private"
)

// DefaultFrom is the default from address for all outgoing emails
var DefaultFrom = &mail.Address{
	Name:    "townsourced",
	Address: "noreply@townsourced.com",
}

var client *sg.SGClient

// Init initialized the mail package, creates any need clients
func Init(testMode bool) error {
	//create email api client
	if !testMode {
		client = sg.NewSendGridClientWithApiKey(private.SendGridAPIKey)
	}

	return nil
}

// Send sends an email From and To Addresses are automatically included in the template data
func Send(from, to *mail.Address, subject, body string) error {
	//send email

	mail := sg.NewMail()
	mail.AddRecipient(to)
	mail.SetFromEmail(from)
	mail.SetSubject(subject)
	mail.SetHTML(body)

	// don't wait for email and dont' return email errors to end users
	if client != nil {
		go func() {
			err := client.Send(mail)

			if err != nil {
				log.WithField("To", to).
					WithField("subject", subject).
					WithField("body", body).
					Errorf("Error sending email: %s", err)
			}
		}()
	}

	return nil
}
