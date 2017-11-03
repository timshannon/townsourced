// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"crypto/tls"
	"fmt"
	"time"

	rt "git.townsourced.com/townsourced/gorethink"
	log "git.townsourced.com/townsourced/logrus"
)

var tables []*table

// DatabaseName is the name of the townsourced database
const DatabaseName = "townsourced"

// DBConfig is database config
type DBConfig struct {
	Address   string   `json:"address,omitempty"`
	Addresses []string `json:"addresses,omitempty"`
	Database  string   `json:"database,omitempty"`
	AuthKey   string   `json:"authkey,omitempty"`
	Timeout   string   `json:"timeout,omitempty"`
	timeout   time.Duration
	TLSConfig *tls.Config `json:"tlsconfig,omitempty"`

	MaxIdle int `json:"max_idle,omitempty"`
	MaxOpen int `json:"max_open,omitempty"`

	// DiscoverHosts is used to enable host discovery, when true the driver
	// will attempt to discover any new nodes added to the cluster and then
	// start sending queries to these new nodes.
	DiscoverHosts bool `json:"discoverHosts,omitempty"`
	// NodeRefreshInterval is used to determine how often the driver should
	// refresh the status of a node.
	NodeRefreshInterval time.Duration `json:"nodeRefreshInterval,omitempty"`
}

// TODO: Handle shard and replication at this level?
var session *rt.Session

func prepDB() error {
	for i := range tables {
		err := tables[i].ensure()
		if err != nil {
			return err
		}
	}

	return nil
}

func ensureDB(dbname string) error {
	var dbNames []string
	c, err := rt.DBList().Run(session)
	if err != nil {
		return err
	}
	err = c.All(&dbNames)
	if err != nil {
		return err
	}

	if !inStr(dbNames, dbname) {

		log.Debugf("Creating Database %s", dbname)
		_, err := rt.DBCreate(dbname).RunWrite(session)
		if err != nil {
			return err
		}
	}
	return nil
}

func inStr(slice []string, value string) bool {
	for i := range slice {
		if slice[i] == value {
			return true
		}
	}
	return false
}

type table struct {
	rt.Term
	name     string
	database string
	rt.TableCreateOpts
	indexes []index
}

type index struct {
	name string
	rt.IndexCreateOpts
	indexFunc interface{}
	table     *table
}

func (t *table) ensure() error {
	if t.database == "" {
		t.database = DatabaseName
	}
	err := ensureDB(t.database)
	if err != nil {
		return err
	}
	var tables []string
	db := rt.DB(t.database)
	c, err := db.TableList().Run(session)
	if err != nil {
		return err
	}

	err = c.All(&tables)
	if err != nil {
		return err
	}

	if !inStr(tables, t.name) {
		if t.TableCreateOpts.PrimaryKey == nil {
			t.TableCreateOpts.PrimaryKey = "Key"
		}
		log.Debugf("Creating Table %s", t.name)
		_, err = db.TableCreate(t.name, t.TableCreateOpts).RunWrite(session)
		if err != nil {
			return err
		}
	}

	t.Term = rt.DB(t.database).Table(t.name)

	for i := range t.indexes {
		t.indexes[i].table = t
		err = t.indexes[i].ensure()
		if err != nil {
			return err
		}
	}

	// wait for all indexes to finish building
	_, err = t.IndexWait().Run(session)
	if err != nil {
		return err
	}
	return nil
}

func (i *index) ensure() error {
	var indexes []string

	t := rt.DB(i.table.database).Table(i.table.name)

	c, err := t.IndexList().Run(session)
	if err != nil {
		return err
	}
	err = c.All(&indexes)
	if err != nil {
		return err
	}

	if !inStr(indexes, i.name) {
		log.Debugf("Creating Index %s on Table %s", i.name, i.table.name)
		if i.indexFunc != nil {
			_, err = t.IndexCreateFunc(i.name, i.indexFunc, i.IndexCreateOpts).RunWrite(session)
		} else {
			_, err = t.IndexCreate(i.name, i.IndexCreateOpts).RunWrite(session)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func wErr(res rt.WriteResponse, err error) error {
	if err == rt.ErrEmptyResult {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if res.Errors > 0 {
		return fmt.Errorf("Rethink write error: count %d, First Error: %s", res.Errors, res.FirstError)
	}
	return nil
}

// DatabaseSession returns the underlying rethinkdb database session
// should usually only be used in tools and tests
func DatabaseSession() *rt.Session {
	return session
}
