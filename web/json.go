// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	log "git.townsourced.com/townsourced/logrus"
	"github.com/timshannon/townsourced/fail"
)

const (
	statusSuccess = "success"
	statusError   = "error"
	statusFail    = "fail"
)

const maxJSONSize = 1 << 20 //10MB

var errInputTooLarge = &fail.Fail{
	Message:    "Input size is too large, please check your input and try again",
	HTTPStatus: http.StatusRequestEntityTooLarge,
}

// JSend is the standard format for a response from townsourced
type JSend struct {
	Status   string      `json:"status"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
	Failures []error     `json:"failures,omitempty"`
	More     bool        `json:"more,omitempty"` // more data exists for this request
}

type etagger interface {
	Etag() string
}

//respondJsend marshalls the input into a json byte array
// and writes it to the reponse with appropriate header
func respondJsend(w http.ResponseWriter, response *JSend) {
	respondJsendCode(w, response, 0)
}

// respondJsendCode is the same as respondJSend, but lets you specify a status code
func respondJsendCode(w http.ResponseWriter, response *JSend, statusCode int) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json")

	if len(response.Failures) > 0 && response.Message == "" {
		response.Message = "One or more item has failed. Check the individual failures for details."
	}

	result, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		log.Errorf("Error marshalling response: %s", err)

		result, _ = json.Marshal(&JSend{
			Status:  statusError,
			Message: "An internal error occurred, and we'll look into it.",
		})
	}

	if statusCode <= 0 {
		switch response.Status {
		case statusFail:
			w.WriteHeader(http.StatusBadRequest)
		case statusError:
			w.WriteHeader(http.StatusInternalServerError)
		}
		//default is status 200
	} else {
		w.WriteHeader(statusCode)
	}

	// if data has an etag use it
	switch t := response.Data.(type) {
	case etagger:
		w.Header().Set("ETag", t.Etag())
	}

	_, err = w.Write(result)
	if err != nil {
		log.WithField("JSEND", result).Errorf("Error writing jsend response: %s", err)
	}
}

func parseInput(r *http.Request, result interface{}) error {
	//TODO: use sync.Pool of buffers

	lr := &io.LimitedReader{R: r.Body, N: maxJSONSize + 1}
	buff, err := ioutil.ReadAll(lr)
	if err != nil {
		return err
	}

	if lr.N == 0 {
		return errInputTooLarge
	}

	if len(buff) == 0 {
		return nil
	}

	err = json.Unmarshal(buff, result)
	if err != nil {
		return err
	}
	return nil
}
