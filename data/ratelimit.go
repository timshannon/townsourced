// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import "time"

// Request Attempts will exist entirely in cache
// they will automatically expire if there hasn't been a new
// attempt in the request attempt time range

type cacheRateLimit struct {
	id          string
	requestType string
	timeRange   time.Duration
}

func (r *cacheRateLimit) key() string {
	return r.id + "_" + r.requestType
}

func (r *cacheRateLimit) source(result interface{}) error {
	return nil
}

func (r *cacheRateLimit) refresh() {
	//nothing to refresh
	return
}

func (r *cacheRateLimit) dependents() []cacher {
	return nil
}

func (r *cacheRateLimit) expiration() time.Duration {
	return r.timeRange
}

// AttemptsGet gets all previous attempts for the given ipaddress+type
func AttemptsGet(result interface{}, id, requestType string) error {
	return cacheGet(&cacheRateLimit{
		id:          id,
		requestType: requestType,
	}, result)
}

// AttemptsSet sets the current attemps for the given ipaddress / type
func AttemptsSet(attempts interface{}, id, requestType string, timeRange time.Duration) error {
	return cacheSet(&cacheRateLimit{
		id:          id,
		requestType: requestType,
		timeRange:   timeRange,
	}, attempts)
}
