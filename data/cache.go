// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"git.townsourced.com/townsourced/consistent"
	"git.townsourced.com/townsourced/gomemcache/memcache"
	log "git.townsourced.com/townsourced/logrus"
)

//TODO: Log statistics on query usage to determine
// which queries should become cache entries

// CacheConfig is cache server connection configuration
type CacheConfig struct {
	Addresses []string `json:"addresses,omitempty"`
}

type cacher interface {
	key() string
	source(result interface{}) error
	expiration() time.Duration
	refresh()             // refresh the cache with no data input
	dependents() []cacher // list of dependent cachers that need to be updated when this one is updated
}

var cacheClient *memcache.Client

// cacheGet tries to retrieve data from cache for the passed in definition
// if not found in cache, it'll retrieve the data from the database and update
// the cache with the retrieved value
func cacheGet(c cacher, result interface{}) error {
	item, err := cacheClient.Get(c.key())
	//get cache based on c.key()
	if err == nil {
		//key found
		return cacheDecode(item.Value, result)
	}

	log.WithField("key", c.key()).Debugf("Cache Miss: %s", err)
	// if not found, get data from c.source()
	err = c.source(result)
	if err != nil {
		return err
	}
	if result == nil {
		return ErrNotFound
	}

	return cacheSet(c, result)
}

// cacheSet will update the cache value for the passed in cacher definition
func cacheSet(c cacher, value interface{}) error {
	if value == nil {
		return cacheClient.Delete(c.key())
	}

	cacheValue, err := cacheEncode(value)
	if err != nil {
		return err
	}

	err = cacheClient.Set(&memcache.Item{
		Key:        c.key(),
		Value:      cacheValue,
		Expiration: int32(c.expiration().Seconds()),
	})
	if err != nil {
		return err
	}

	deps := c.dependents()
	for i := range deps {
		go func(cr cacher) {
			cr.refresh()
		}(deps[i])
	}

	return nil
}

// cache refresh is different from cacheSet in that the value is not passed in
// just the value type
func cacheRefresh(c cacher, emptyType interface{}) error {
	err := c.source(emptyType)
	if err != nil {
		return err
	}

	return cacheSet(c, emptyType)
}

func initCache(cfg *CacheConfig) error {
	selector, err := newConsistentSelector(cfg.Addresses...)
	if err != nil {
		return err
	}

	cacheClient = memcache.NewFromSelector(selector)
	return nil
}

type consistentSelector struct {
	mu    sync.RWMutex
	con   *consistent.Consistent
	addrs map[string]net.Addr
}

//TODO: Add / remove cache servers on the fly
// automatically drop out failing cache servers
// Track healthy cache servers in the DB, poll for new ones every 15 minutes

func newConsistentSelector(servers ...string) (*consistentSelector, error) {
	cs := &consistentSelector{
		con:   consistent.New(),
		addrs: make(map[string]net.Addr),
	}

	for _, server := range servers {
		if strings.Contains(server, "/") {
			addr, err := net.ResolveUnixAddr("unix", server)
			if err != nil {
				return nil, err
			}
			cs.addrs[addr.String()] = addr
		} else {
			tcpaddr, err := net.ResolveTCPAddr("tcp", server)
			if err != nil {
				return nil, err
			}
			cs.addrs[tcpaddr.String()] = tcpaddr
		}
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()
	for i := range cs.addrs {
		cs.con.Add(cs.addrs[i].String())
	}
	return cs, nil
}

// Each iterates over each server calling the given function
func (cs *consistentSelector) Each(f func(net.Addr) error) error {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	for _, a := range cs.addrs {
		if err := f(a); nil != err {
			return err
		}
	}
	return nil
}

// PickServer picks a cache server to use
func (cs *consistentSelector) PickServer(key string) (net.Addr, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	addr, err := cs.con.Get(key)
	if err != nil {
		return nil, err
	}
	naddr, ok := cs.addrs[addr]
	if !ok {
		return nil, memcache.ErrNoServers
	}
	return naddr, nil
}

// cacheEncode encodes the passed in value for storage in cache
// currently this means gzipped json
// TODO: Potential issue with struct tags if I have something I
// don't want to be sent to the client, but I want stored in cache
// currently I'm just manually clearing it at the app level,which
// I'm ok with
func cacheEncode(value interface{}) ([]byte, error) {
	//TODO: use sync.Pool of buffers and gzip writers
	result := bytes.NewBuffer(make([]byte, 0, 4096))
	w := gzip.NewWriter(result)

	err := json.NewEncoder(w).Encode(value)
	if err != nil {
		return nil, fmt.Errorf("Error encoding cache value: %s", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("Error encoding cache value: %s", err)
	}

	return result.Bytes(), nil
}

// cacheDecode decodes cache values
func cacheDecode(data []byte, v interface{}) (err error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err == io.EOF {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error decoding cache value: %s", err)
	}
	defer func() {
		if cerr := r.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	err = json.NewDecoder(r).Decode(v)
	if err != nil {
		return fmt.Errorf("Error decoding cache value: %s", err)
	}

	return nil
}

// CacheClient returns the underlying memcached client
// should usually only be used in tools and tests
func CacheClient() *memcache.Client {
	return cacheClient
}
