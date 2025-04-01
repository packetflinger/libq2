package demo

import (
	"encoding/hex"
	"testing"

	pb "github.com/packetflinger/libq2/proto"
)

func TestMvdMarshalServerData(t *testing.T) {
	tests := []struct {
		name string
		data *pb.MvdServerData
		want string
	}{
		{
			name: "test0",
			data: &pb.MvdServerData{
				VersionMajor: 37,
				VersionMinor: 2010,
				SpawnCount:   123456789,
				GameDir:      "uranus",
				ClientNumber: 20,
			},
			want: "25000000da0715cd5b077572616e7573001400",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := NewMVD2Writer(&pb.MvdDemo{})
			msg := writer.MarshalServerData(tc.data)
			got := hex.EncodeToString(msg.Data)
			if got != tc.want {
				t.Errorf("MarshalServerData(%v) = %s, want %s\n", tc.data, got, tc.want)
			}
		})
	}
}
