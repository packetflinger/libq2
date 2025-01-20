package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Server
		wantErr bool
	}{
		{
			name:    "empty",
			input:   "",
			want:    Server{},
			wantErr: true,
		},
		{
			name:    "no port",
			input:   "1.2.3.4",
			want:    Server{},
			wantErr: true,
		},
		{
			name:  "IPv4",
			input: "1.2.3.4:27999",
			want: Server{
				Address: "1.2.3.4",
				Port:    27999,
			},
			wantErr: false,
		},
		{
			name:  "hostname with port",
			input: "test.srv.addr:27910",
			want: Server{
				Address: "test.srv.addr",
				Port:    27910,
			},
			wantErr: false,
		},
		{
			name:  "IPv6 with port",
			input: "[2003:db8::3]:27988",
			want: Server{
				Address: "2003:db8::3",
				Port:    27988,
			},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewServer(tc.input)
			if err != nil && !tc.wantErr {
				t.Errorf("NewServer(%v) returned unexpected error: %v\n", tc.input, err)
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("NewServer(%v) diff: %v\n", tc.input, diff)
			}
		})
	}
}
