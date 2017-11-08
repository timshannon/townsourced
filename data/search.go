// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"git.townsourced.com/townsourced/elastic"
	log "git.townsourced.com/townsourced/logrus"
)

var searchClient *elastic.Client
var searchTypes []*searchType

// SearchConfig is search server connection configuration
type SearchConfig struct {
	Addresses  []string          `json:"addresses,omitempty"`
	MaxRetries int               `json:"maxRetries,omitEmpty"`
	Index      SearchIndexConfig `json:"index,omitempty"`
}

// SearchIndexConfig is config for how the main elasticSearch index should be configure
type SearchIndexConfig struct {
	Name     string `json:"name,omitempty"`
	Shards   int    `json:"shards,omitempty"`
	Replicas int    `json:"replicas,omitempty"`
}

func initSearch(cfg *SearchConfig) error {
	var err error
	searchClient, err = elastic.NewClient(
		elastic.SetURL(cfg.Addresses...),
		elastic.SetMaxRetries(cfg.MaxRetries),
		elastic.SetHealthcheckTimeoutStartup(30*time.Second),
	)
	if err != nil {
		return err
	}

	log.Debugf("Connected to Search Instance")

	//prep indexes

	exists, err := searchClient.IndexExists(cfg.Index.Name).Do()
	if err != nil {
		return err
	}

	if !exists {
		log.Debugf("Prepping Search Index")
		result, err := searchClient.CreateIndex(cfg.Index.Name).BodyJson(map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   cfg.Index.Shards,
				"number_of_replicas": cfg.Index.Replicas,
			},
			"mappings": buildSearchMappings(),
		}).Do()
		if err != nil {
			return err
		}

		if !result.Acknowledged {
			return fmt.Errorf("Failed to create elasticsearch index %s", cfg.Index.Name)
		}
	}

	//prep mappings
	indexResult, err := searchClient.IndexGet().Index(cfg.Index.Name).Do()
	if err != nil {
		return err
	}

	currentMappings := indexResult[cfg.Index.Name].Mappings

	for _, m := range searchTypes {
		if _, ok := currentMappings[m.name]; !ok {
			log.Debugf("Adding Search Mapping %s", m.name)
			//add mapping
			resp, err := searchClient.PutMapping().Index(cfg.Index.Name).Type(m.name).
				BodyJson(m.mappingDefinition()).Do()
			if err != nil {
				return err
			}
			if !resp.Acknowledged {
				return fmt.Errorf("Error creating elasticsearch mapping %s", m.name)
			}
		}
		m.indexName = cfg.Index.Name
	}

	return nil
}

func buildSearchMappings() map[string]interface{} {
	mappings := make(map[string]interface{})
	for i := range searchTypes {
		mappings[searchTypes[i].name] = searchTypes[i].mappingDefinition()
	}
	return mappings
}

type searchType struct {
	name       string
	properties map[string]interface{}
	indexName  string
}

func (s *searchType) mappingDefinition() map[string]interface{} {
	return map[string]interface{}{
		"properties": s.properties,
	}
}

// index will update or insert a new document for the given key (upsert)
func (s *searchType) index(key string, value interface{}) error {
	//resp, err := searchClient.Update().Index(s.indexName).Type(s.name).Id(key).BodyJson(value).Do()
	_, err := searchClient.Update().Index(s.indexName).Type(s.name).Id(key).Doc(value).DocAsUpsert(true).Do()
	if err != nil {
		if elErr, ok := err.(*elastic.Error); ok && elErr.Status == http.StatusNotFound {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *searchType) delete(key string) error {
	resp, err := searchClient.Delete().Index(s.indexName).Type(s.name).Id(key).Do()
	if err != nil {
		if elErr, ok := err.(*elastic.Error); ok && elErr.Status == http.StatusNotFound {
			return ErrNotFound
		}
		return err
	}

	if !resp.Found {
		return ErrNotFound
	}

	return nil
}

func (s *searchType) search(query elastic.Query) *elastic.SearchService {
	return searchClient.Search().Index(s.indexName).Query(query)
}

// SearchResult is the result values of a search
//  used for Deserializing data from the results of a search
// TODO: Highlight results?
// TODO: Search Suggestions
type SearchResult struct {
	result *elastic.SearchResult
	index  int
}

// Next fetches the next value from the search result
// returns io.EOF when there are no more items
// example:
//	posts = make([]Post, result.Count())
//
//	for i := range posts {
//		err = result.Next(&posts[i])
//		if err != nil {
//			return nil, err
//		}
//	}
func (r *SearchResult) Next(result interface{}) error {
	if r.result.TotalHits() == 0 {
		return ErrNotFound
	}

	if len(r.result.Hits.Hits) == 0 || r.index >= len(r.result.Hits.Hits) {
		return io.EOF
	}
	err := json.Unmarshal(*r.result.Hits.Hits[r.index].Source, result)

	if err != nil {
		return err
	}
	r.index++

	return nil
}

// Count returns the number of search results
func (r *SearchResult) Count() int {
	return len(r.result.Hits.Hits)
}

// SearchClient returns the underlying elasticSearch client
// should usually only be used in tools and tests
func SearchClient() *elastic.Client {
	return searchClient
}
