// These functions relate to figuring out the physical location of a particular
// player. The current granularity is limited to the country.
//
// This functionlity depends on the IP-to-Country Lite CSV list downloaded
// from: https://db-ip.com/db/download/ip-to-country-lite. Any list will work
// really, but each record is expected to be in the following format:
//
//	x.x.x.x,y.y.y.y,zz
//
// Where x.x.x.x is the first IP in a range, y.y.y.y is the last IP in a range
// and zz is the Alpha-2 ISO 3166 country code.
//
// Note: these are not CIDR ranges and not even expected to be within the
// appropriate byte-boundaries for them. Converting these into actual CIDR
// ranges will in many cases result in a single record resolving to multiple
// CIDR ranges and possibly include some unintended top-end space. Just keep
// them as literal ranges of addresses.
//
// Example 63.254.184.0 to 64.4.0.1 overlaps two different class As and various
// byte boundaries. This range would resolve to:
//
//	[
//	 63.254.184.0/21,
//	 63.254.192.0/18,
//	 63.255.0.0/16,
//	 64.0.0.0/14,
//	 64.4.0.0/31,
//	]
//
// Not only is this a mega-bitch to caculate it'll result in 4 extra lookups to
// match an IP. Don't even get me started on doing this for IPv6 🤮
package state

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/packetflinger/libq2/message"
)

const (
	GZipMagic = 35615 // 0x1f 0x8b
)

// A mapping of an IP range to a particular country. We can't use a *net.IPNet
// because first and last addresses aren't necessarily in the same CIDR range.
//
// Supports IPv4 and IPv6
type GeoIP struct {
	First   net.IP
	Last    net.IP
	Country string // 2 character country code
}

// Make the list a type for use as a func receiver.
type GeoIPList []GeoIP

// Read each line of the ip-to-country database file into memory. This file
// should be in the format:
func LoadGeoIPFile(inputFile string) (GeoIPList, error) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return GeoIPList{}, fmt.Errorf("GeoIP file error: %v", err)
	}
	return LoadGeoIPData(data)
}

// Create the list of GeoIPs from raw bytes. Also determines if the data is
// compressed and handles that situation transparently. Only Gzip compression
// is supported.
func LoadGeoIPData(data []byte) (GeoIPList, error) {
	var out []GeoIP
	var err error

	// not even enough data to tell what type it is
	if len(data) < 4 {
		return out, fmt.Errorf("geoip data not specified")
	}

	buf := message.NewBuffer(data[0:3])
	if buf.ReadWord() == GZipMagic {
		data, err = gzippedContent(data)
		if err != nil {
			return out, fmt.Errorf("error decompressing GeoIP data: %v", err)
		}
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
		out = append(out, GeoIP{
			First:   net.ParseIP(tokens[0]),
			Last:    net.ParseIP(tokens[1]),
			Country: strings.ToLower(tokens[2]),
		})
	}
	return out, nil
}

// The data parameter is assumed to be deflated using gzip. This decompresses
// it in memory and returns the inflated bytes.
func gzippedContent(data []byte) ([]byte, error) {
	var out []byte
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return out, fmt.Errorf("reading gzip format: %v", err)
	}
	defer gzReader.Close()
	content, err := io.ReadAll(gzReader)
	if err != nil {
		return out, fmt.Errorf("reading compressed data: %v", err)
	}
	return content, nil
}

// Get the country code for a particular IP address string. Output of "zz" is
// considered country-less (think RFC1918). This is also used for any address
// not found in the list.
func (n *GeoIPList) Lookup(addr string) string {
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
