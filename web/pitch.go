// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"regexp"
)

// pitch handler is a special handler that loads a different page based on the user agent string
// it wraps a standard staticHandler
type pitchHandler struct {
	standard *staticHandler
	simple   *staticHandler
}

// Earlier Iphones appear to crash trying to load our pitch page, this regexp matches against all IOS and serves
// them up a simplier version of the pitch page
var rxIOS = regexp.MustCompile(`^(?:(?:(?:Mozilla/\d\.\d\s*\()+|Mobile\s*Safari\s*\d+(?:\.\d+)+\s*)(?:iPhone(?:\s+Simulator)?|iPad|iPod);\s*(?:U;\s*)?(?:[a-z]+(?:-[a-z]+)?;\s*)?CPU\s*(?:iPhone\s*)?(?:OS\s*\d+_\d+(?:_\d+)?\s*)?(?:like|comme)\s*Mac\s*O?S?\s*X(?:;\s*[a-z]+(?:-[a-z]+)?)?\)\s*)?(?:AppleWebKit/\d+(?:\.\d+(?:\.\d+)?|\s*\+)?\s*)?(?:\(KHTML,\s*(?:like|comme)\s*Gecko\s*\)\s*)?(?:(?:Version|CriOS)/\d+(?:\.\d+)+\s*)?(?:Mobile/\w+\s*)?(?:Safari/\d+(?:\.\d+)*.*)?$`)

func (p *pitchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ua := r.UserAgent()

	if rxIOS.MatchString(ua) {
		// serve up simple pitch page
		p.simple.ServeHTTP(w, r)
		return
	}
	p.standard.ServeHTTP(w, r)
}
