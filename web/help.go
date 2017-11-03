// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"

	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
)

func helpTemplate(w http.ResponseWriter, r *http.Request, c context) {
	key := data.NewKey(c.params.ByName("key"))

	help, err := app.HelpGet(key)
	if errHandledPage(err, w, r, c) {
		return
	}

	err = w.(*templateWriter).execute("HELP", help)

	if err != nil {
		log.Errorf("Error executing help template: %s", err)
	}
}
