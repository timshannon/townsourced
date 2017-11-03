// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/app"
)

func adminTemplate(w http.ResponseWriter, r *http.Request, c context) {
	// ?since=<since>

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()

	if errHandledPage(err, w, r, c) {
		return
	}

	if !u.Admin {
		four04(w, r)
		return
	}

	values := r.URL.Query()

	since := time.Time{}

	if values.Get("since") != "" {
		since, err = time.Parse(time.RFC3339, values.Get("since"))
		if err != nil {
			since = time.Time{}
		}
	}

	stats, err := app.AdminStatsGet(u, since)

	err = w.(*templateWriter).execute("ADMIN", struct {
		Stats *app.AdminStats
		Error error
	}{
		Stats: stats,
		Error: err,
	})
	if err != nil {
		log.Errorf("Error executing admin template: %s", err)
	}
}
