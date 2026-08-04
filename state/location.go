// These functions relate to figuring out the physical location of a particular
// player. The current granularity is limited to the country.
package state

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

// A mapping of an IP range to a particular country. We can't use a *net.IPNet
// because first and last addresses aren't necessarily in the same CIDR range.
// For example 63.254.184.0 to 64.4.0.1 will overlaps two different class A
// ranges.
//
// Supports IPv4 and IPv6
type NetCountry struct {
	First   net.IP
	Last    net.IP
	Country string // 2 character country code
}

// Make the list a type for use as a func receiver
type NetCountryList []NetCountry

// Read each line of the ip-to-country database file into memory. This file
// should be in the format:
//
//	x.x.x.x,y.y.y.y,zz
//
// where x.x.x.x is the first IP in the range, y.y.y.y is the last IP in the
// range, and zz is the country code.
//
// The `cont` parameter configures whether or not to continue if a parse error
// occurs or to give up and return an error message.
func LoadIPCountries(inputFile string, cont bool) (NetCountryList, error) {
	var out []NetCountry
	if inputFile == "" {
		return out, fmt.Errorf("country database not specified")
	}
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return out, fmt.Errorf("reading country db: %s", err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "0.0.0.0") {
			// 0.0.0.0 will match everything, just don't include it
			continue
		}
		tokens := strings.Split(line, ",")
		if len(tokens) != 3 {
			continue
		}
		out = append(out, NetCountry{
			First:   net.ParseIP(tokens[0]),
			Last:    net.ParseIP(tokens[1]),
			Country: strings.ToLower(tokens[2]),
		})
	}
	return out, nil
}

// Get the country code for a particular IP address string. Output of "zz" is
// considered country-less (think RFC1918).
func (n *NetCountryList) Country(addr string) string {
	if addr == "" {
		return "zz"
	}
	ip := net.ParseIP(addr)
	for _, nc := range *n {
		c1 := bytes.Compare(nc.First, ip)
		c2 := bytes.Compare(nc.Last, ip)
		if (c1 == 0 || c1 == -1) && (c2 == 0 || c2 == 1) {
			return nc.Country
		}
	}
	return "zz"
}
