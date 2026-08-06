package state

import (
	"testing"
)

func TestCountry(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{
			name: "localhost",
			addr: "127.0.0.5",
			want: "zz",
		},
		{
			name: "rfc1918-1",
			addr: "10.3.5.6",
			want: "zz",
		},
		{
			name: "pf chicago server",
			addr: "169.197.131.131",
			want: "us",
		},
		{
			name: "pf germany server",
			addr: "86.105.53.128",
			want: "de",
		},
		{
			name: "pf au server",
			addr: "103.73.64.180",
			want: "au",
		},
	}

	IPCountryList, err := LoadIPCountries("/tmp/ipcountry.csv.gz")
	if err != nil {
		t.Fatalf("unable to load country db: %v", err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := (&IPCountryList).Lookup(tc.addr)
			if got != tc.want {
				t.Errorf("Country(%q) = %q, want %q", tc.addr, got, tc.want)
			}
		})
	}
}
