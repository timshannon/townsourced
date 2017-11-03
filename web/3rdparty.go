// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"

	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/fail"
)

//TODO: Get rid of this, make token at request to
// specific endpoint and build oauth urls on server side, rather than client side
// client should only ever see our 3rd party token, and never 3rd party oauth credentials
// even if temporary, this pretty low on the priority list though

func thirdPartyGet(w http.ResponseWriter, r *http.Request, c context) {
	token := r.URL.Query().Get("token")

	state, err := app.ThirdPartyStateGet(token)
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   state,
	})
}

func thirdPartyPost(w http.ResponseWriter, r *http.Request, c context) {
	input := &app.ThirdPartyState{}

	if errHandled(parseInput(r, input), w, r, c) {
		return
	}

	if input.Provider != "facebook" && input.Provider != "google" && input.Provider != "twitter" {
		errHandled(fail.New("Invalid 3rd party provider", input), w, r, c)
		return
	}

	token, err := app.ThirdPartyStateNew(input.ReturnURL, input.Provider)
	if errHandled(err, w, r, c) {
		return
	}
	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   token,
	})
	return
}
