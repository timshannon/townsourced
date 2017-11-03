// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"time"

	"git.townsourced.com/townsourced/elastic"
	rt "git.townsourced.com/townsourced/gorethink"
	log "git.townsourced.com/townsourced/logrus"
)

// AnnouncementTown is the key of the townsourced announcements town
const AnnouncementTown = "announcements"

func init() {
	tables = append(tables, tblTown)
	searchTypes = append(searchTypes, srcTown)
}

var tblTown = &table{
	name: "town",
	indexes: []index{
		index{
			name: "Location",
			IndexCreateOpts: rt.IndexCreateOpts{
				Geo:   "Location",
				Multi: nil,
			},
		},
		index{name: "Created"},
	},
}

var srcTown = &searchType{
	name: "town",
	properties: map[string]interface{}{
		//"_all": map[string]interface{}{ // issue with elasticsearch 2.2?
		//"enabled": false,
		//},
		"name": map[string]interface{}{
			"type":     "string",
			"analyzer": "snowball",
		},
		"description": map[string]interface{}{
			"type":     "string",
			"analyzer": "snowball",
		},
	},
}

// Towns returns a set of towns in the database
func Towns(result interface{}, keys ...Key) (err error) {
	ikeys := make([]interface{}, len(keys))
	for i := range keys {
		ikeys[i] = keys[i]
	}
	c, err := tblTown.GetAll(ikeys...).Run(session)

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

// TownGet retrieves a single town
func TownGet(result interface{}, key Key) error {
	return cacheGet(&cacheTown{townKey: key}, result)
}

// TownGetByLocation retrieves all towns within the distance passed from the location passed in
func TownGetByLocation(result interface{}, locationQry LocationSearcher, from, limit int) error {
	c, err := locationQry.query(tblTown, "Location", limit).Filter(map[string]interface{}{
		"Private": false,
	}).Filter(rt.Row.Field("Key").Eq(AnnouncementTown).Not()).Skip(from).Run(session)

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

// TownGetBySearch retrieves a list of towns based on a full text search of the town's name and description
func TownGetBySearch(search string, from, limit int) (*SearchResult, error) {
	//qry := elastic.NewBoolQuery().Should(elastic.NewMatchQuery(search, "name")).
	//Should(elastic.NewMatchQuery(search, "description")).MinimumNumberShouldMatch(1)

	result, err := srcTown.search(elastic.NewMultiMatchQuery(search, "name", "description")).
		Sort("_score", false).
		From(from).
		Size(limit).
		Do()

	if err != nil {
		return nil, err
	}

	if result.TotalHits() == 0 {
		return nil, ErrNotFound
	}

	return &SearchResult{
		result: result,
		index:  0,
	}, nil
}

// TownInsert inserts a new town
func TownInsert(data interface{}, key Key) error {
	err := wErr(tblTown.Insert(data).RunWrite(session))
	if err != nil {
		return err
	}

	ct := &cacheTown{townKey: key, data: data}
	ct.refresh()

	return nil
}

// TownUpdate updates a town
func TownUpdate(data interface{}, key Key) error {
	err := tryUpdateVersion(tblTown.Get(key), data)
	if err != nil {
		return err
	}
	ct := &cacheTown{townKey: key, data: data}
	ct.refresh()

	return nil
}

// single town cache
type cacheTown struct {
	townKey Key
	data    interface{}
}

func (t *cacheTown) key() string {
	return "town_" + string(t.townKey)
}

func (t *cacheTown) source(result interface{}) (err error) {
	c, err := tblTown.Get(t.townKey).Run(session)
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

	return c.One(result)
}

func (t *cacheTown) dependents() []cacher {
	return nil
}

func (t *cacheTown) expiration() time.Duration {
	return time.Duration(0)
}

func (t *cacheTown) refresh() {
	if t.data == nil {
		err := cacheRefresh(t, map[string]interface{}{})
		if err != nil {
			log.Errorf("Error refreshing town cache. Error: %s", err)
		}
		return
	}

	err := cacheSet(t, t.data)
	if err != nil {
		log.Errorf("Error refreshing town cache. Error: %s", err)
	}
}

// TownIndex indexes the town for full text searching
func TownIndex(town interface{}, key Key) error {
	return srcTown.index(string(key), town)
}

// TownRemoveIndex removes the given town from the full text search indexes
func TownRemoveIndex(key Key) error {
	return srcTown.delete(string(key))
}

// TownAllCount returns the count of the total number of towns, usually used by maintenance and not the frontend
func TownAllCount() (int, error) {
	c, err := tblTown.Count().Run(session)
	if err != nil {
		return -1, err
	}
	defer func() {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	result := 0
	err = c.One(&result)
	if err != nil {
		return -1, err
	}
	return result, nil
}

// TownGetAll retrieves all towns
// This likely shouldn't be used for the actual website, and should only be used for maintenance / tasks
func TownGetAll(result interface{}, from, limit int) error {
	c, err := tblTown.Skip(from).Limit(limit).Run(session)
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

// town population cache
type cacheTownPopulation struct {
	townKey Key
}

func (t *cacheTownPopulation) key() string {
	return "town_" + string(t.townKey) + "_population"
}

func (t *cacheTownPopulation) source(result interface{}) error {
	c, err := tblUser.Filter(func(user rt.Term) rt.Term {
		return user.Field("TownKeys").Contains(func(tk rt.Term) rt.Term {
			return tk.Field("Key").Eq(t.townKey)
		})
	}).Count().Run(session)

	if err != nil {
		return err
	}

	defer func() {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if c.IsNil() {
		return nil
	}

	return c.One(result)
}

func (t *cacheTownPopulation) dependents() []cacher {
	return nil
}

func (t *cacheTownPopulation) expiration() time.Duration {
	return 30 * time.Second
}

func (t *cacheTownPopulation) refresh() {
	pop := 0

	err := t.source(&pop)
	if err != nil {
		log.Errorf("Error refreshing town population cache. Error: %s", err)
	}

	err = cacheSet(t, pop)
	if err != nil {
		log.Errorf("Error refreshing town population cache. Error: %s", err)
	}
}

// TownGetPopulation retrieves the current population of a town
func TownGetPopulation(key Key) (int, error) {
	pop := 0
	err := cacheGet(&cacheTownPopulation{townKey: key}, &pop)
	if err != nil {
		return 0, err
	}

	return pop, nil
}
