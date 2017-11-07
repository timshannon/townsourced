// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

// Package data is the package for dealing with the data layer of townsourced.  This means the following:
//  	RethinkDB
//	Memcache
//	Elasticsearch
//
// NOTE: For Rethinkdb, be very careful when using maps in structs, as updates won't remove fields in maps
//  You'll need to due a full replace instead.  Generally it's safer, and more performant to use a slice instead
package data

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	rt "git.townsourced.com/townsourced/gorethink"
	log "git.townsourced.com/townsourced/logrus"
)

// ErrNotFound is the error returned when no records are found
// for the given criteria
var ErrNotFound = errors.New("Data not found")

// EmptyKey is an empty key value
const EmptyKey = Key("")

// MaxKeyLength is the longest possible Key
const MaxKeyLength = 127

// Key is a unique identifier for a given peice of data
// 	Key handling rules are as follows
//		Never use data.Key(value), use data.NewKey(value)
//		Web layer code deals with strings, app layer code deals with keys
//			the transfer from Web to App will make sure keys are properly cased
type Key string

// NewKey returns a new Key from the passed in string
// use this instead of data.Key(key) to make new keys to ensure that
// it's properly lowercase
func NewKey(s string) Key {
	if len(s) > MaxKeyLength {
		s = s[:MaxKeyLength]
	}
	return Key(strings.ToLower(s))
}

// UnmarshalJSON implments a custom JSON unmarshaller for reading a Key from JSON input, to ensure all keys
// coming from user input via JSON are properly lower case
func (k *Key) UnmarshalJSON(input []byte) error {
	str := ""
	err := json.Unmarshal(input, &str)
	if err != nil {
		return err
	}

	*k = NewKey(str)
	return nil
}

// KeyWhen returns a new KeyWhen struct
func (k Key) KeyWhen() KeyWhen {
	return KeyWhen{
		Key:  k,
		When: time.Now(),
	}
}

// NewKeySlice is a help for building a slice of keys from strings
func NewKeySlice(s []string) []Key {
	keys := make([]Key, len(s))
	for i := range keys {
		keys[i] = NewKey(s[i])
	}
	return keys
}

// KeyWhen is a helper for the serveral instances where you want to keep track of a list
// of joining keys as well as when those keys where added
// joining a town
// saved posts, etc
type KeyWhen struct {
	Key  Key       `json:"key"`
	When time.Time `json:"when"`
}

// KeyWhenSlice is a slice of KeyWhens
type KeyWhenSlice []KeyWhen

// Keys gets a list of keys form a KeyWhen slice
func (k KeyWhenSlice) Keys() []Key {
	keys := make([]Key, len(k))
	for i := range keys {
		keys[i] = k[i].Key
	}
	return keys
}

// A UUID is like a key (a unique identifier), but it is machine generated rather than user generated
// and it is shortened for user facing urls and json data
type UUID string

// EmptyUUID is an empty UUID value
const EmptyUUID = UUID("")

// NewUUID returns a new generated UUID
// this should rarely be used, and the database should mostly be responsible for generating new UUIDs
func NewUUID() UUID {
	f, err := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	if err != nil {
		panic("New UUID cannot be generated because /dev/urandom is inaccessible.")
	}
	b := make([]byte, 16)
	_, err = f.Read(b)
	if err != nil {
		panic("New UUID cannot be generated because /dev/urandom is inaccessible.")
	}
	f.Close()
	return UUID(fmt.Sprintf("%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
}

// ToUUID takes a Base64 encoded string and converts it to a full UUID
// if it can't be decoded then an empty UUID is returned
func ToUUID(b64 string) UUID {
	if len(b64) == 36 {
		//assume already UUID
		return UUID(b64)
	}
	g, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil || len(g) != 16 {
		return EmptyUUID
	}

	return UUID(fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", g[:4], g[4:6], g[6:8], g[8:10], g[10:]))
}

// FromUUID creates a url friendly shortend string from the passed in UUID
func FromUUID(g UUID) string {
	if len(g) != 36 {
		return string(g)
	}
	if g[8] != '-' || g[13] != '-' || g[18] != '-' || g[23] != '-' {
		return string(g)
	}
	uuid := make([]byte, 16)
	for i, x := range []int{
		0, 2, 4, 6, 9, 11, 14, 16, 19, 21, 24, 26, 28, 30, 32, 34} {
		v, ok := hexToByte(string(g[x:]))
		if !ok {
			return string(g)
		}
		uuid[i] = v
	}

	return base64.RawURLEncoding.EncodeToString(uuid)
}

// UnmarshalJSON implments a custom JSON unmarshaller for reading the base64 encoded value that the UUID was
// written to
func (g *UUID) UnmarshalJSON(input []byte) error {
	str := ""
	err := json.Unmarshal(input, &str)
	if err != nil {
		return err
	}

	*g = ToUUID(str)

	return nil
}

// MarshalJSON implements a custom JSON marshaler for a UUID to shorten the UUID via base64 for more readable URLs
func (g *UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(FromUUID(*g))
}

// UUIDWhen returns a new UUIDWhen struct
func (g UUID) UUIDWhen() UUIDWhen {
	return UUIDWhen{
		Key:  g,
		When: time.Now(),
	}
}

// UUIDWhen is a helper for the serveral instances where you want to keep track of a list
// of joining UUIDS as well as when those guids where added
// saved posts, etc
type UUIDWhen struct {
	Key  UUID      `json:"key"`
	When time.Time `json:"when"`
}

// UUIDWhenSlice is a slice of UUIDWhens
type UUIDWhenSlice []UUIDWhen

// UUIDs gets a list of guids form a UUIDWhen slice
func (g UUIDWhenSlice) UUIDs() []UUID {
	guids := make([]UUID, len(g))
	for i := range guids {
		guids[i] = g[i].Key
	}
	return guids
}

// Config is data layer configuration
// thinks like database ip addresses, cache servers, etc
type Config struct {
	DB      DBConfig     `json:"db"`
	Cache   CacheConfig  `json:"cache"`
	Search  SearchConfig `json:"search"`
	DevMode bool         `json:"-"`
}

// DefaultConfig returns the default configuration for the data layer
func DefaultConfig() *Config {
	return &Config{
		DB: DBConfig{
			Address:  "127.0.0.1:28015",
			Database: DatabaseName,
			Timeout:  "60s",
		},
		Cache: CacheConfig{
			Addresses: []string{"127.0.0.1:11211"},
		},
		Search: SearchConfig{
			Addresses:  []string{"http://127.0.0.1:9200"},
			MaxRetries: 0,
			Index: SearchIndexConfig{
				Name:     "townsourced",
				Shards:   5,
				Replicas: 1,
			},
		},
	}
}

// Init initialized the data layer based on the passed in
// configuration
func Init(cfg *Config) error {
	var err error

	rt.SetVerbose(cfg.DevMode)
	cfg.DB.timeout, err = time.ParseDuration(cfg.DB.Timeout)
	if err != nil {
		return fmt.Errorf("Error parsing DB Timeout: %s", err)
	}

	rtConnect(cfg)

	log.WithField("CFG", cfg.DB).Debugf("Connected to DB")

	err = initCache(&cfg.Cache)
	if err != nil {
		return err
	}

	log.WithField("CFG", cfg.Cache).Debugf("Cache Initialized")

	err = initSearch(&cfg.Search)
	if err != nil {
		return err
	}

	log.WithField("CFG", cfg.Search).Debugf("Search Initialized")

	err = prepDB()

	log.Debugf("DB Prepped")

	return err
}

func rtConnect(cfg *Config) {
	var err error
	session, err = rt.Connect(rt.ConnectOpts{
		Address:             cfg.DB.Address,
		Addresses:           cfg.DB.Addresses,
		Database:            cfg.DB.Database,
		AuthKey:             cfg.DB.AuthKey,
		Timeout:             cfg.DB.timeout,
		TLSConfig:           cfg.DB.TLSConfig,
		MaxIdle:             cfg.DB.MaxIdle,
		MaxOpen:             cfg.DB.MaxOpen,
		DiscoverHosts:       cfg.DB.DiscoverHosts,
		NodeRefreshInterval: cfg.DB.NodeRefreshInterval,
	})

	if err != nil {
		log.Warnf("Error connecting to database: %s  RETRYING...", err)
		time.Sleep(5 * time.Second)
		rtConnect(cfg)
	}
}
