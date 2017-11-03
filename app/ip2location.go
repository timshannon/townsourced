// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package app

import "github.com/timshannon/townsourced/data"

const ip2locationTownDistance = townSearchMaxDistance // distance to search for towns

// IPLocation is the data structure from the IP2Location csv DB
type IPLocation struct {
	IPFrom      uint64  `json:"-"`
	IPTo        uint64  `json:"-"`
	CountryCode string  `json:"-"`
	RegionName  string  `json:"region,omitempty"`
	CityName    string  `json:"city,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

// IPToLocation returns a location for the passed in IP Address
func IPToLocation(ipAddress string) (*IPLocation, error) {
	result := &IPLocation{}

	err := data.IP2LocationGet(result, ipAddress)
	if err != nil {
		if err == data.ErrNotFound {
			return result, nil
		}
		return nil, err
	}

	return result, nil
}

// IPToTowns returns the nearest registered towns to the given IP address
func IPToTowns(ipAddress string, limit int) ([]Town, error) {
	location, err := IPToLocation(ipAddress)
	if err != nil {
		return nil, err
	}

	if location == nil {
		return nil, nil
	}

	towns, err := TownSearchDistance(location.Longitude, location.Latitude, ip2locationTownDistance, 0, limit)
	if err != nil {
		return nil, err
	}

	return towns, nil
}
