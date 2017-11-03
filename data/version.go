// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"git.townsourced.com/townsourced/townsourced/fail"

	rt "git.townsourced.com/townsourced/gorethink"
)

//TODO: Might be excessive, but we could cache all current version / etag values, and
// setup the head web response to read from cache, so all etag checks would be very
// cheap

// ErrVersionStale is when an update is attempted with a version older than the current version
// in the database
var ErrVersionStale = &fail.Fail{
	Message: "You are trying to update this record based on old information.  " +
		"Please refresh your screen and try again.",
	HTTPStatus: http.StatusConflict,
}

// Used for version checks to make sure updates don't apply
// to the wrong version of a record
type versioner interface {
	Ver() string      //current version of the data
	VerField() string // location of the version info
	Rev()             // Generate a new version
}

// Version is meant to be used as an anonymous struct to
// gain built in version protection on updates, and track created an updated dates
type Version struct {
	VerTag  string    `json:"vertag,omitempty" gorethink:",omitempty"`
	Created time.Time `json:"created,omitempty" gorethink:",omitempty"`
	Updated time.Time `json:"updated,omitempty" gorethink:",omitempty"`
}

// Rev generates a new unique version identifier
func (v *Version) Rev() {
	if v.VerTag == "" {
		//New version, set created time
		v.Created = time.Now()
	}
	v.Updated = time.Now()
	// for testing
	//v.Updated = time.Date(2015, time.Month(random.Intn(12-1)+10), random.Intn(30), random.Intn(23), random.Intn(59), random.Intn(59), 0, v.Updated.Location())

	bits := 64
	result := make([]byte, bits/8)
	_, err := io.ReadFull(rand.Reader, result)
	if err != nil {
		panic(fmt.Sprintf("Error generating random values: %v", err))
	}

	b64 := base64.StdEncoding.EncodeToString(result)
	//make url safe
	v.VerTag = strings.TrimRight(strings.Replace(strings.Replace(b64, "+", "-", -1), "/", "_", -1), "=")
}

// Ver returns the unique version of this record
func (v *Version) Ver() string {
	return v.VerTag
}

// VerField returns the field in which the version is stored
func (v *Version) VerField() string {
	return "VerTag"
}

func tryUpdateVersion(selection rt.Term, data interface{}) error {
	if v, ok := data.(versioner); ok {
		current := v.Ver()
		v.Rev()
		w, err := selection.Update(rt.Branch(rt.Row.Field(v.VerField()).Eq(current), data, nil)).
			RunWrite(session)
		err = wErr(w, err)
		if err != nil {
			return err
		}

		if w.Replaced == 0 {
			//Check for version mismatch by checking
			// if selection matches records without version
			// If nothing is found, then errnotfound
			// if something is found, then version mismatch
			c, err := selection.Run(session)
			if err != nil {
				return err
			}
			if c.IsNil() {
				return ErrNotFound
			}
			return ErrVersionStale
		}
		return nil
	}
	return wErr(selection.Update(data).RunWrite(session))
}
