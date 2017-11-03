// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
)

const postPageLimit = 40

// root page handler
func rootTemplate(w http.ResponseWriter, r *http.Request, c context) {
	if demoMode {
		err := w.(*templateWriter).execute("DEMO", nil)
		if err != nil {
			log.Errorf("Error executing DEMO template: %s", err)
		}
		return
	}

	if c.session == nil {
		location, err := app.IPToLocation(ipAddress(r))
		if errHandledPage(err, w, r, c) {
			return
		}

		local, towns, err := app.PostGetByLocation(location.Longitude, location.Latitude, 300, "", time.Time{}, 6)
		if errHandledPage(err, w, r, c) {
			return
		}

		if len(local) < 6 {
			towns, err = app.IPToTowns(ipAddress(r), 6)
			if errHandledPage(err, w, r, c) {
				return
			}
		}

		err = w.(*templateWriter).execute("PUBLIC", struct {
			Local []app.Post
			Towns []app.Town
		}{
			Local: local,
			Towns: towns,
		})
		if err != nil {
			log.Errorf("Error executing public template: %s", err)
		}
		return
	}

	u, err := c.session.User()

	if errHandledPage(err, w, r, c) {
		return
	}

	posts, err := app.PostGetByTowns(u, data.KeyWhenSlice(u.TownKeys).Keys(), "", time.Time{}, postPageLimit, false)
	if errHandledPage(err, w, r, c) {
		return
	}

	var towns []app.Town
	townHint := false

	if len(posts) < postPageLimit || len(u.TownKeys) < 2 {
		townHint = true
		localTowns, err := app.IPToTowns(ipAddress(r), 10)
		if errHandledPage(err, w, r, c) {
			return
		}

		// add only towns that they aren't already a member of
		for i := range localTowns {
			found := false
			for j := range u.TownKeys {
				if u.TownKeys[j].Key == localTowns[i].Key {
					found = true
				}
			}
			if !found {
				towns = append(towns, localTowns[i])
			}
		}
	}

	err = w.(*templateWriter).execute("ROOT", struct {
		Posts    []app.Post
		User     *app.User
		TownHint bool
		Towns    []app.Town
	}{
		Posts:    posts,
		User:     u,
		TownHint: townHint,
		Towns:    towns,
	})

	if err != nil {
		log.Errorf("Error executing ROOT template: %s", err)
	}
}
