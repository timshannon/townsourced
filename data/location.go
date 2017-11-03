// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"fmt"

	rt "git.townsourced.com/townsourced/gorethink"
	"git.townsourced.com/townsourced/gorethink/types"
)

/*
	LocationUnit defines units of measure for location based queries
*/
const (
	LocationUnitMeter        = "m"
	LocationUnitKilometer    = "km"
	LocationUnitMile         = "mi"
	LocationUnitNauticalMile = "nm"
	LocationUnitFoot         = "ft"
)

// LatLng is a Latitude and Longitude grouping
type LatLng types.Point

// NewLatLng returns a new LatLng type for use with locations
func NewLatLng(latitude, longitude float64) (LatLng, error) {
	ll := LatLng(types.Point{})

	if longitude < -180 || longitude > 180 {
		return ll, fmt.Errorf("Invalid longitude %f", longitude)
	}
	ll.Lon = longitude
	if latitude < -90 || latitude > 90 {
		return ll, fmt.Errorf("Invalid latitude %f", latitude)
	}

	ll.Lat = latitude
	return ll, nil
}

func validateUnit(unit string) {
	if unit != LocationUnitFoot && unit != LocationUnitKilometer && unit != LocationUnitMeter &&
		unit != LocationUnitMile && unit != LocationUnitNauticalMile {
		panic(fmt.Sprintf("Invalid location query unit %s", unit))
	}
}

// LocationSearcher is an interface for searching based on location
type LocationSearcher interface {
	query(t *table, index string, limit int) rt.Term
}

// DistanceSearch is a location search based on distance from a single point
type DistanceSearch struct {
	point       types.Point
	latitude    float64
	longitude   float64
	unit        string
	maxDistance float64
}

// NewDistanceSearch creates a new distance search type
func NewDistanceSearch(latLng LatLng, maxDistance float64, unit string) *DistanceSearch {
	validateUnit(unit)
	return &DistanceSearch{
		point:       types.Point(latLng),
		unit:        unit,
		maxDistance: maxDistance,
	}
}

func (d *DistanceSearch) query(t *table, index string, limit int) rt.Term {
	return t.GetNearest(d.point, rt.GetNearestOpts{
		Index:      index,
		Unit:       d.unit,
		MaxResults: limit,
		MaxDist:    d.maxDistance,
	}).Map(func(val rt.Term) rt.Term {
		return val.Field("doc")
	})
}

// AreaSearch is a location search for everything withing a rectangle area
type AreaSearch struct {
	nw LatLng
	ne LatLng
	sw LatLng
	se LatLng
}

// NewAreaSearch returns a new area search query
func NewAreaSearch(northBounds, southBounds, eastBounds, westBounds float64) (*AreaSearch, error) {
	if northBounds == southBounds {
		return nil, fmt.Errorf("North and South latitudes cannot be equal")
	}

	if eastBounds == westBounds {
		return nil, fmt.Errorf("East and West longitudes cannot be equal")
	}

	if northBounds < southBounds {
		southBounds, northBounds = northBounds, southBounds
	}

	if eastBounds < westBounds {
		westBounds, eastBounds = eastBounds, westBounds
	}

	nw, err := NewLatLng(northBounds, westBounds)
	if err != nil {
		return nil, err
	}

	ne, err := NewLatLng(northBounds, eastBounds)
	if err != nil {
		return nil, err
	}

	sw, err := NewLatLng(southBounds, westBounds)
	if err != nil {
		return nil, err
	}
	se, err := NewLatLng(southBounds, eastBounds)
	if err != nil {
		return nil, err
	}

	//check for wraparound
	if northBounds == 90 && southBounds == -90 {
		return nil, fmt.Errorf("North and South latitudes cannot be equal")
	}

	if eastBounds == 180 && westBounds == -180 {
		return nil, fmt.Errorf("East and West longitudes cannot be equal")
	}

	return &AreaSearch{
		nw: nw,
		ne: ne,
		sw: sw,
		se: se,
	}, nil
}

func (a *AreaSearch) query(t *table, index string, limit int) rt.Term {
	return t.GetIntersecting(rt.Polygon(types.Point(a.nw),
		types.Point(a.ne),
		types.Point(a.se),
		types.Point(a.sw)), rt.GetIntersectingOpts{
		Index: index,
	})
}
