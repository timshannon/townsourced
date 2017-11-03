// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"time"

	rt "git.townsourced.com/townsourced/gorethink"
)

func init() {
	tables = append(tables, tblTask)
}

const taskDatabase = "task"

var tblTask = &table{
	name:     "task",
	database: taskDatabase,
	indexes: []index{
		index{
			name: "Owner",
			indexFunc: func(row rt.Term) interface{} {
				return []interface{}{row.Field("Owner"), row.Field("Closed")}
			},
		},
		index{name: "Type"},
	},
}

// TaskInsert inserts a new task to be run
func TaskInsert(task interface{}) error {
	return wErr(tblTask.Insert(task).RunWrite(session))
}

// TaskUpdate updates a single task
func TaskUpdate(task interface{}, key UUID) error {
	return wErr(tblTask.Get(key).Update(task).RunWrite(session))
}

// TaskClaim marks the next set of unclaimed, non-closed tasks as owned by the given user
// this is to immediately prevent any other task runners from sharing these tasks
// A task should only belong to one runner at a time
func TaskClaim(owner Key, limit uint) error {
	return wErr(tblTask.GetAllByIndex("Owner", []interface{}{EmptyKey, false}).
		Filter(rt.Row.Field("NextRun").Le(time.Now())).OrderBy("Priority", "Created").
		Limit(limit).Update(map[string]interface{}{
		"Owner": owner,
	}).RunWrite(session))
}

// TaskGetMine retrieves all unprocessed tasks that have been marked as the owenres, who's nextRun time has passed
// in order by priority, then oldest tasks
func TaskGetMine(result interface{}, owner Key) (err error) {
	c, err := tblTask.GetAllByIndex("Owner", []interface{}{owner, false}).OrderBy("Priority", "Created").Run(session)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	if c.IsNil() {
		return ErrNotFound
	}
	return c.All(result)
}

// TaskGetOpenType gets all open tasks of a given type
func TaskGetOpenType(result interface{}, taskType string, limit uint) (err error) {
	c, err := tblTask.GetAllByIndex("Type", taskType).Limit(limit).Run(session)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	if c.IsNil() {
		return ErrNotFound
	}
	return c.All(result)
}

// TaskDeleteClosed deletes all closed tasks
func TaskDeleteClosed() error {
	return wErr(tblTask.GetAllByIndex("Owner", []interface{}{EmptyKey, true}).
		Delete(rt.DeleteOpts{Durability: "soft"}).RunWrite(session))
}
