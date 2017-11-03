// Townsourced

package web

import (
	"bytes"
	"net"
	"net/http"
	"strings"
)

// Copyright 2016 Tim Shannon. All rights reserved.

//ipAddress returns the actual ip address from the request
func ipAddress(r *http.Request) string {
	// list of possible addresses from request header, from most to least internal until we get a public address
	addresses := append(strings.Split(r.Header.Get("X-Forwarded-For"), ","),
		strings.Split(r.Header.Get("X-Real-Ip"), ",")...)

	for i := range addresses {
		ip := addresses[i]
		// header can contain spaces too, strip those out.
		realIP := net.ParseIP(strings.Replace(ip, " ", "", -1))

		if !realIP.IsGlobalUnicast() && !isPrivateSubnet(realIP) {
			// bad address, go to next
			continue
		}

		return realIP.String()
	}

	// if none found use remoteAddress as last resort
	index := strings.LastIndex(r.RemoteAddr, ":")
	if index == -1 {
		return r.RemoteAddr
	}
	return r.RemoteAddr[:index]
}

//ipRange - a structure that holds the start and end of a range of ip addresses
type ipRange struct {
	start net.IP
	end   net.IP
}

// inRange - check to see if a given ip address is within a range given
func (r ipRange) in(ipAddress net.IP) bool {
	// strcmp type byte comparison
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) <= 0 {
		return true
	}
	return false
}

var privateRanges = []ipRange{
	ipRange{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	ipRange{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	ipRange{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	ipRange{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	ipRange{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	ipRange{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}

// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		for i := range privateRanges {
			if privateRanges[i].in(ipAddress) {
				return true
			}
		}
	}

	return false
}
