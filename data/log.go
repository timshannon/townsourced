// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	rt "git.townsourced.com/townsourced/gorethink"
)

const logDatabase = "log"

func init() {
	tables = append(tables, tblLog)
}

var tblLog = &table{
	name:     "log",
	database: logDatabase,
	indexes: []index{
		index{name: "Time"},
	},
}

// Log writes a new log entry
func Log(entry interface{}) error {
	return wErr(tblLog.Insert(entry, rt.InsertOpts{
		Durability:    "soft",
		ReturnChanges: false,
	}).RunWrite(session))
}
