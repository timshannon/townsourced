// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import log "git.townsourced.com/townsourced/logrus"

// Halt cleanups the townsourced app, and shuts it down as clean as
// possible logging the passed in message, and printing it to the stderr
func Halt(msg string, a ...interface{}) {
	//TODO: shutdown cleanup, logging, etc
	stopTaskRunner()
	log.Fatalf(msg+"\n", a...)
}
