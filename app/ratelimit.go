// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

//TODO: Simplify Rate limiting, and provide headers https://developer.github.com/v3/#rate-limiting

// RequestAttempt is a client request of some type
// Login attempt, town submission, user creation, etc
// Used to prevent rapid requests / attempts
type RequestAttempt struct {
	ID   string // ID is a unique identifier for the requestor, e.g. IP address, sessionID, or API key
	Type RequestType
	When time.Time
}

// RequestType is a type of request that needs to be rate limited
type RequestType struct {
	Type         string
	FreeAttempts int           //Number of attempts before delay occurs
	Scale        time.Duration //attempt # * scale = delay
	Range        time.Duration // Range to look for previous requests
	MaxWait      time.Duration //Max wait time that can be accumulated
}

//Defaults for request attempts
// By Default any given request can be made attemptsFree times with no delay
// After attemptsFree, requests will start getting delayed attemptScale seconds for
// each number of requests past attemptsFree in the past attemptsRange minutes
// If maxWait is reached, then an error will immediately be returned
const (
	attemptsFree   = 500
	attemptScale   = 1 * time.Second
	attemptRange   = 10 * time.Second
	attemptMaxWait = 30 * time.Second
	attemptType    = "general"
)

// ErrRequestMax is when too many requests are attempted
var ErrRequestMax = &fail.Fail{
	Message:    "Too many requests are being made.",
	HTTPStatus: 429,
}

func (t *RequestType) setDefault() {
	if t.Type == "" {
		t.Type = attemptType
	}
	if t.FreeAttempts == 0 {
		t.FreeAttempts = attemptsFree
	}
	if t.Scale == 0 {
		t.Scale = attemptScale
	}
	if t.Range == 0 {
		t.Range = attemptRange
	}
	if t.MaxWait == 0 {
		t.MaxWait = attemptMaxWait
	}
}

// AttemptRequest logs a new attempt for the given ip address / or session ID of the given type, and waits
// appropriately if they are exceeding the rate limit for the type
func AttemptRequest(id string, reqType RequestType) error {
	var attempts []RequestAttempt
	reqType.setDefault()

	err := data.AttemptsGet(&attempts, id, reqType.Type)
	if err != nil && err != data.ErrNotFound {
		return err
	}

	attempts = append([]RequestAttempt{RequestAttempt{
		ID:   id,
		Type: reqType,
		When: time.Now(),
	}}, attempts...)

	//newer attempts are tacked onto the front so once the first attempt outside if the range is found
	// all following will also be outside the range
	maxIndex := len(attempts)
	for i := range attempts {
		if attempts[i].When.Before(time.Now().Add(-1 * reqType.Range)) {
			maxIndex = i
			break
		}
	}
	attempts = attempts[:maxIndex]

	err = data.AttemptsSet(attempts, id, reqType.Type, reqType.Range)
	if err != nil {
		return err
	}

	if len(attempts) > reqType.FreeAttempts {
		wait := time.Duration(int64(len(attempts)-reqType.FreeAttempts) * int64(reqType.Scale))
		if wait > reqType.MaxWait {
			return ErrRequestMax
		}

		log.WithField("ID", id).Debugf("Start Ratelimit waiting for %s", reqType.Type)
		time.Sleep(wait)
		log.WithField("ID", id).Debugf("Finish Ratelimit waiting for %s", reqType.Type)
	}

	return nil
}
