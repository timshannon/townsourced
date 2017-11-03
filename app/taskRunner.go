// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"sync"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/data"
)

/*
	Tasks get run by task runners in the following way
	1. Taskrunners update a queue of ready to process tasks to run as owned by them (the hostname of the
		given server)
	2. The Taskrunner selects everything the database that's marked as theirs and processes them
	3. Any failures get thrown back into the queue by updating the owner as none ""
	4. Anything completed and not set to run again gets marked as closed and never run again

	Some tasks may be recurring and continually put themselves back into the queue for processing at a later time
	Some tasks may be one off and close once they have run once
*/

var taskRun = make(chan bool)

func startTaskRunner(owner data.Key, queueSize uint, pollInterval time.Duration) {
	var wg sync.WaitGroup

	tasks := make([]*Task, 0, queueSize)
	go func() {
		taskRun <- true
	}()

	for <-taskRun {
		err := data.TaskClaim(owner, queueSize)
		if err != nil {
			log.Errorf("Error claiming open tasks from DB: %s", err)
			continue
		}

		err = data.TaskGetMine(&tasks, owner)
		if err != nil {
			log.Errorf("Error getting open owned tasks from DB: %s", err)
			continue
		}

		for i := range tasks {
			wg.Add(1)
			go func(t *Task) {
				t.Run()
				wg.Done()
			}(tasks[i])
		}
		wg.Wait()
		time.AfterFunc(pollInterval, func() {
			taskRun <- true

		})
	}
}

func stopTaskRunner() {
	taskRun <- false
}
