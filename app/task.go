// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/data"
)

const (
	priorityHigh = iota + 1
	priorityMediumHigh
	priorityMedium
	priorityMediumLow
	priorityLow
)

var tasks *taskMap

type taskMap struct {
	sync.RWMutex
	t map[string]Tasker
}

func initTasks(owner data.Key, queueSize uint, pollInterval time.Duration) {
	tasks = &taskMap{
		t: make(map[string]Tasker),
	}

	recurringTasks := []Tasker{
		&deleteClosedTasker{},
		&taskerUnusedImages{},
	}

	registerRecurringTask(recurringTasks)

	go startTaskRunner(owner, queueSize, pollInterval)
}

func registerTaskType(t Tasker) {
	tasks.Lock()
	defer tasks.Unlock()

	tasks.t[t.Type()] = t
}

// registerRecurringTask is for tasks that regularly recur and aren't one-off tasks
// if the task isn't already in the task DB, then it inserts it
func registerRecurringTask(recurringTasks []Tasker) {

	for i := range recurringTasks {
		var tasks []*Task
		t := recurringTasks[i]
		err := data.TaskGetOpenType(&tasks, t.Type(), 1)

		registerTaskType(t)
		if err == nil && len(tasks) >= 1 {
			//task already exists in the DB
			continue
		}

		if err != nil && err != data.ErrNotFound {
			panic(fmt.Sprintf("Unable to register recurring task %s: %s", t.Type(), err))
		}

		//not found or len == 0, insert new task entry
		err = taskAdd(t, nil)
		if err != nil {
			panic("Unable to register recurring task " + t.Type())
		}
	}
}

func (t *taskMap) get(Type string) Tasker {
	t.RLock()
	defer t.RUnlock()

	tskr, ok := t.t[Type]
	if !ok {
		panic("Unregistered Task Type!: " + Type)
	}
	return tskr
}

// Tasker is a unit of work that can be scheduled or queued to run later by the task runner
// tasks should be atomic operations, and if a task fails, it should leave no paritially commited data behind
type Tasker interface {
	Type() string                      // Type of the task
	Priority() uint                    // Task Priority so high priority tasks (1) will be run before low priority tasks (5)
	NextRun() time.Time                // detrmines when to next run this task, a zero time means the task is complete, and is not run again
	Do(variables ...interface{}) error // The task to be run
	Retry() int                        // Number of times to retry this task before marking it as failed, return -1 will retry forever
}

// Task is a unit of work stored in the database, corresponds to a pre-registered tasker interface
type Task struct {
	Key       data.UUID `gorethink:",omitempty"`
	Type      string    `gorethink:",omitempty"`
	Owner     data.Key
	Priority  uint          `gorethink:",omitempty"`
	NextRun   time.Time     `gorethink:",omitempty"`
	Variables []interface{} `gorethink:",omitempty"`
	Created   time.Time     `gorethink:",omitempty"`
	Failed    time.Time     `gorethink:",omitempty"`
	Completed time.Time     `gorethink:",omitempty"`
	Closed    bool
	Retry     int

	tasker Tasker
}

// taskAdd adds a new task
// Note that if variables are complex structs they need to be serializable into and out of the database
// best practice is to keep them a slice of base types
func taskAdd(t Tasker, variables ...interface{}) error {
	if t.Priority() < priorityHigh || t.Priority() > priorityLow {
		return errors.New("Invalid task priority")
	}

	next := t.NextRun()
	// all tasks must run at least once
	if next.IsZero() {
		next = time.Now()
	}

	task := &Task{
		Type:      t.Type(),
		Priority:  t.Priority(),
		NextRun:   next,
		Variables: variables,
		Created:   time.Now(),
		Retry:     0,
	}

	return data.TaskInsert(task)
}

// Run runs the given task
// exported in case this needs to be run from an external process
func (t *Task) Run() {
	t.tasker = tasks.get(t.Type)

	if t.errHandled(t.tasker.Do(t.Variables...)) {
		return
	}

	t.NextRun = t.tasker.NextRun()
	t.Retry = 0             //reset any retries
	t.Owner = data.EmptyKey // throw it back into the queue

	if t.NextRun.IsZero() {
		t.Closed = true
		t.Completed = time.Now()
	}

	err := data.TaskUpdate(t, t.Key)
	if err != nil {
		log.WithFields(log.Fields{
			"taskKey": t.Key,
			"type":    t.Type,
		}).Errorf("An error occured when completing a task: %s", err)
	}
}

func (t *Task) errHandled(err error) bool {
	if err == nil {
		return false
	}

	log.WithFields(log.Fields{
		"taskKey": t.Key,
		"type":    t.Type,
	}).Errorf("An error occured running a task: %s", err)

	//Check and set retry
	t.fail()
	return true
}

// fail either marks the tasks as failed and closed if retry limit is reached
// or increments the retry limit, and queues it up to be run again
func (t *Task) fail() {
	if t.Retry >= t.tasker.Retry() && t.tasker.Retry() != -1 {
		t.Failed = time.Now()
		t.Closed = true
	} else {
		t.Retry++
	}

	t.Owner = data.EmptyKey //throw it back in the queue for processing by any available task runner

	err := data.TaskUpdate(t, t.Key)
	if err != nil {
		log.WithFields(log.Fields{
			"taskKey": t.Key,
			"type":    t.Type,
		}).Errorf("An error occured when updating a task for failure: %s", err)
	}
}

//Delete Tasker

type deleteClosedTasker struct{}

func (d *deleteClosedTasker) Type() string       { return "DeleteClosedTasks" }
func (d *deleteClosedTasker) Priority() uint     { return priorityLow }
func (d *deleteClosedTasker) NextRun() time.Time { return time.Now().Add(15 * time.Minute) }
func (d *deleteClosedTasker) Retry() int         { return -1 }
func (d *deleteClosedTasker) Do(variables ...interface{}) error {
	err := data.TaskDeleteClosed()
	if err == data.ErrNotFound {
		return nil
	}
	return err
}
