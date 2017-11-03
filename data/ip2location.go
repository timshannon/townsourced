// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package data

import (
	"encoding/binary"
	"net"

	rt "git.townsourced.com/townsourced/gorethink"
)

func init() {
	tables = append(tables, tblIP2Location)
}

const externalDatabase = "external" // data from outside of townsourced

var tblIP2Location = &table{
	name:     "ip2location",
	database: externalDatabase,
	TableCreateOpts: rt.TableCreateOpts{
		PrimaryKey: "IPFrom",
	},
}

// IP2LocationGet retrieves a LatLng location from the passed in IPAddress
func IP2LocationGet(result interface{}, ipAddress string) error {
	iNum := IPNumber(ipAddress)

	c, err := tblIP2Location.Between(rt.MinVal, iNum,
		rt.BetweenOpts{
			RightBound: "closed",
		}).OrderBy("IPFrom", rt.OrderByOpts{
		Index: rt.Desc("IPFrom"),
	}).Limit(1).Filter(rt.Row.Field("IPTo").Ge(iNum)).Run(session)

	if err != nil {
		return err
	}

	if c.IsNil() {
		return ErrNotFound
	}

	c.One(result)

	return nil
}

// IP2LocationTruncate truncates the IP2Location table by dropping it and recreating it.  Much faster than
// deleting all the records individually
func IP2LocationTruncate() error {
	err := wErr(rt.DB(tblIP2Location.database).TableDrop(tblIP2Location.name).RunWrite(session))
	if err != nil {
		return err
	}

	return tblIP2Location.ensure()
}

// IP2LocationImport imports an array for IP2Location entries
func IP2LocationImport(entries interface{}) error {
	return wErr(tblIP2Location.Insert(entries).RunWrite(session))
}

// IPNumber returns a sortable int version of a string IP Address
func IPNumber(address string) uint64 {
	ip := net.ParseIP(address)

	v4 := ip.To4()
	if v4 == nil {
		//ipv6
		v6 := ip.To16()
		num := binary.BigEndian.Uint64([]byte(v6[8:]))
		num <<= 64
		num += binary.BigEndian.Uint64([]byte(v6))
		return uint64(num)
	}

	return uint64(binary.BigEndian.Uint32([]byte(v4)))
}
