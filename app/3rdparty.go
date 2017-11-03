// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"net/http"
	"time"

	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

// ThirdPartyState is used both
// for getting the user back to where they were when they logged in, as well as
// CSRF protection
type ThirdPartyState struct {
	Token     string `json:"token" gorethink:",omitempty"`
	ReturnURL string `json:"returnURL" gorethink:",omitempty"`
	Provider  string `json:"provider" gorethink:",omitempty"`
}

// ThirdPartyStateNew generates a new 3rdparty state tracking token.
func ThirdPartyStateNew(returnURL, provider string) (string, error) {
	state := &ThirdPartyState{
		Token:     Random(256),
		ReturnURL: returnURL,
		Provider:  provider,
	}

	err := data.TempTokenSet(state, state.Token, 15*time.Minute)
	if err != nil {
		return "", err
	}
	return state.Token, nil
}

// ThirdPartyStateGet retrieves the 3rdparty user state from the passed in token
func ThirdPartyStateGet(token string) (*ThirdPartyState, error) {
	state := &ThirdPartyState{}

	err := data.TempTokenGet(state, token)
	if err == data.ErrNotFound {
		return nil, &fail.Fail{
			Message:    "State not found for Token",
			Data:       token,
			HTTPStatus: http.StatusNotFound,
		}
	}

	if err != nil {
		return nil, err
	}
	return state, nil
}
