// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

// rate limit the number of new posts that can be created by the same user
var shareRequestType = app.RequestType{
	Type:         "share",
	FreeAttempts: 10,
	Scale:        5 * time.Second,
	Range:        10 * time.Minute,
	MaxWait:      1 * time.Minute,
}

func shareTemplate(w http.ResponseWriter, r *http.Request, c context) {
	// ?url=<url>
	// ?title=<title>
	// ?content=<content>
	// ?image=<imageURL>
	// ?image=<imageURL>
	// ?image=<imageURL>
	// ?imagekey=<imageKey>
	// ?imagekey=<imageKey>
	// ?imagekey=<imageKey>
	// ?town=<townKey>
	// ?selector=<cssSelector to parse>

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandledPage(err, w, r, c) {
		return
	}

	//Rate limit shares
	if errHandledPage(app.AttemptRequest(string(c.session.UserKey), shareRequestType), w, r, c) {
		return
	}

	values := r.URL.Query()

	var imageKeys []data.UUID

	for _, k := range values["imagekey"] {
		imageKeys = append(imageKeys, data.ToUUID(k))
	}

	post, err := app.Share(u, values.Get("url"), values.Get("title"), values.Get("content"), r.UserAgent(),
		data.Key(values.Get("town")), values["image"], imageKeys, values.Get("selector"))

	shareError := false

	if err != nil {
		if !fail.IsFail(err) {
			errHandledPage(err, w, r, c)
			return
		}
		shareError = true
	}

	err = w.(*templateWriter).execute("EDITPOST", struct {
		Post       *app.Post
		User       *app.User
		ShareError bool
	}{
		Post:       post,
		User:       u,
		ShareError: shareError,
	})
	if err != nil {
		log.Errorf("Error executing editpost template: %s", err)
	}

}
