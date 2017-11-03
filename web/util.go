// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

func sinceLimitValues(values url.Values, defaultLimit int) (time.Time, int, error) {
	since := values.Get("since")

	t := time.Time{}
	var err error
	if since != "" {
		t, err = time.Parse(time.RFC3339, since)
		if err != nil {
			return time.Time{}, -1, fail.New("Invalid time format for since parameter", since)
		}
	}

	limit, err := strconv.Atoi(values.Get("limit"))
	if err != nil {
		limit = defaultLimit
	}

	return t, limit, nil
}

func uuidGet(w http.ResponseWriter, r *http.Request, c context) {
	w.Header().Set("Content-Type", `text/plain; charset="UTF-8"`)
	_, err := w.Write([]byte(data.ToUUID(c.params.ByName("uuid"))))
	if errHandled(err, w, r, c) {
		return
	}
}
