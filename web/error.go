// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	rt "git.townsourced.com/townsourced/gorethink"
	log "git.townsourced.com/townsourced/logrus"
	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

const (
	acceptHTML = "text/html"
)

// Err404 is a standard 404 error response
var Err404 = errors.New("Resource not found")

var four04Handler *staticHandler
var errorHandler *staticHandler
var unauthorizedHandler *staticHandler

func errHandled(err error, w http.ResponseWriter, r *http.Request, c context) bool {
	if err == nil {
		return false
	}

	var status, errMsg string

	errMsg = err.Error()

	switch err.(type) {

	case *fail.Fail:
		status = statusFail
		if fail.IsEqual(err, Err404) {
			respond404JSON(w, err.(*fail.Fail).Data.(string))
			return true
		}

		log.WithFields(log.Fields{
			"url":     r.URL,
			"method":  r.Method,
			"header":  r.Header,
			"params":  c.params,
			"session": c.session,
			"data":    err.(*fail.Fail).Data,
		}).Warning(err)
	case *http.ProtocolError, *json.SyntaxError, *json.UnmarshalTypeError:
		//Hardcoded external errors which can bubble up to the end users
		// without exposing internal server information, make them failures
		err = fail.NewFromErr(err)
		status = statusFail

		log.WithFields(log.Fields{
			"url":     r.URL,
			"method":  r.Method,
			"header":  r.Header,
			"params":  c.params,
			"session": c.session,
		}).Warning(err)
		errMsg = fmt.Sprintf("We had trouble parsing your input, please check your input and try again: %s", err)
	default:
		if err == data.ErrNotFound || err == rt.ErrEmptyResult {
			// catch all, but hopefull specific app seconds will handle
			// their data not found errors separately
			respond404JSON(w, "")
			return true
		}

		status = statusError
		log.WithFields(log.Fields{
			"url":     r.URL,
			"method":  r.Method,
			"header":  r.Header,
			"params":  c.params,
			"session": c.session,
		}).Error(err)

		if !devMode {
			errMsg = "An internal server error has occurred"
		}
	}

	if status == statusFail {
		respondJsendCode(w, &JSend{
			Status:  status,
			Message: errMsg,
			Data:    err.(*fail.Fail).Data,
		}, err.(*fail.Fail).HTTPStatus)
	} else {
		respondJsend(w, &JSend{
			Status:  status,
			Message: errMsg,
		})
	}

	return true
}

// errHandledPage is the error handling for page templates.  These type of errors should never be failures, and
// so will instead of 500 errors, so we redirect to the 500 page
func errHandledPage(err error, w http.ResponseWriter, r *http.Request, c context) bool {
	if err == nil {
		return false
	}

	if err == data.ErrNotFound {
		four04(w, r)
		return true
	}

	log.WithFields(log.Fields{
		"params":  c.params,
		"session": c.session,
	}).Error(err)

	errorHandler.ServeHTTP(w, r)
	return true
}

// four04 is a standard 404 response if request header accepts text/html
// they'll get a 404 page, otherwise a json response
func four04(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, acceptHTML) && r.Method == "GET" {
		respond404Page(w, r)
		return
	}
	respond404JSON(w, r.URL.String())
}

// unauthorized is when a user tries to access a page that exists, but they
// aren't logged it and can't access it until they do
func unauthorized(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	w.Header().Set("Cache-Control", "no-cache")

	if strings.Contains(accept, acceptHTML) && r.Method == "GET" {
		unauthorizedHandler.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	response := &JSend{
		Status:  statusFail,
		Message: "Unauthorized.  Please re-authenticate and try again.",
		Data:    r.URL.String(),
	}

	w.WriteHeader(http.StatusUnauthorized)
	result, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Error marshalling unauthorized response: %s", err)
		return
	}

	_, err = w.Write(result)
	if err != nil {
		log.Errorf("Error in unauthorized json response: %s", err)
	}
}

func respond404JSON(w http.ResponseWriter, url string) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	response := &JSend{
		Status:  statusFail,
		Message: "Resource not found",
		Data:    url,
	}

	w.WriteHeader(http.StatusNotFound)

	result, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Error marshalling 404 response: %s", err)
		return
	}

	_, err = w.Write(result)
	if err != nil {
		log.Errorf("Error in respond404JSON: %s", err)
	}
}

// four04Page returns a 404 page
func respond404Page(w http.ResponseWriter, r *http.Request) {
	four04Handler.ServeHTTP(w, r)
}

func panicHandler(w http.ResponseWriter, r *http.Request, rec interface{}) {
	if rec != nil {
		if devMode {
			//halt the instance if runtime error, or is running in devmode
			// otherwise log error and try to recover
			buf := make([]byte, 1<<20)
			stack := buf[:runtime.Stack(buf, true)]
			app.Halt("PANIC: %s \n STACK: %s", rec, stack)
		}
		errHandled(fmt.Errorf("townsourced webserver panicked on %v and has recovered", rec), w, r, context{})
		return
	}
}
