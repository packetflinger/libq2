package master

import (
	"encoding/hex"
	"net"
	"strings"
	"testing"

	"github.com/packetflinger/libq2/message"
)

func TestParseMasterResponse(t *testing.T) {
	tests := []struct {
		name    string
		data    string // string of hex values
		want    []MasterClient
		wantErr bool
	}{
		{
			name: "empty",
			data: "",
			want: []MasterClient{},
		},
		{
			name: "1 server",
			data: "a9 c5 83 85 07 d0",
			want: []MasterClient{
				{
					IP:   net.ParseIP("169.197.131.133"),
					Port: 2000,
				},
			},
		},
		{
			name: "2 servers",
			data: "a9 c5 83 85 07 ce 34 06 bd 9f 07 ce",
			want: []MasterClient{
				{
					IP:   net.ParseIP("169.197.131.133"),
					Port: 1998,
				},
				{
					IP:   net.ParseIP("52.6.189.159"),
					Port: 1998,
				},
			},
		},
		{
			name:    "short",
			data:    "a9 c5 83",
			want:    nil,
			wantErr: true,
		},
		{
			name: "1 server plus extra",
			data: "34 06 bd 9f 07 d2 a9 c5 83",
			want: []MasterClient{
				{
					IP:   net.ParseIP("52.6.189.159"),
					Port: 2002,
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// remove spaces and convert string representation of hex to actual hex
			d, _ := hex.DecodeString(strings.ReplaceAll(tc.data, " ", ""))
			buf := message.NewBuffer(d)
			got := ParseMasterResponse(&buf)
			if len(tc.want) != len(got) {
				t.Errorf("Got %d results, want %d\n", len(got), len(tc.want))
			} else {
				for i, cl := range got {
					if cl.IP.String() != tc.want[i].IP.String() {
						t.Errorf("IP Result[%d]: got %s, want %s\n", i, cl.IP.String(), tc.want[i].IP.String())
					}
					if cl.Port != tc.want[i].Port {
						t.Errorf("Port Result[%d]: got %d, want %d\n", i, cl.Port, tc.want[i].Port)
					}
				}
			}
		})
	}
}
