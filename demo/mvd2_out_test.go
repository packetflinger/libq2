package demo

import (
	"encoding/hex"
	"testing"

	pb "github.com/packetflinger/libq2/proto"
)

func TestMvdMarshalServerData(t *testing.T) {
	tests := []struct {
		name string
		demo *pb.MvdDemo
		data *pb.MvdServerData
		want string
	}{
		{
			name: "no demo flags",
			demo: &pb.MvdDemo{},
			data: &pb.MvdServerData{
				VersionMajor: 37,
				VersionMinor: 2010,
				SpawnCount:   123456789,
				GameDir:      "uranus",
				ClientNumber: 20,
			},
			want: "25000000da0715cd5b077572616e7573001400",
		},
		{
			name: "with demo flags",
			demo: &pb.MvdDemo{
				Flags: 715,
			},
			data: &pb.MvdServerData{
				VersionMajor: 37,
				VersionMinor: 2012,
				SpawnCount:   123456789,
				GameDir:      "uranus",
				ClientNumber: 20,
			},
			want: "25000000dc07cb0215cd5b077572616e7573001400",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := NewMVD2Writer(tc.demo)
			msg := writer.MarshalServerData(tc.data)
			got := hex.EncodeToString(msg.Data)
			if got != tc.want {
				t.Errorf("MarshalServerData(%v) = %s, want %s\n", tc.data, got, tc.want)
			}
		})
	}
}

func TestMvdMarshalConfigString(t *testing.T) {
	tests := []struct {
		name  string
		remap *pb.MvdConfigStringRemap
		data  *pb.ConfigString
		want  string
	}{
		{
			name: "test0",
			data: &pb.ConfigString{
				Index: 1800,
				Data:  "bonergarage",
			},
			want: "0807626f6e657267617261676500",
		},
		{
			name: "test1",
			data: &pb.ConfigString{
				Index: 1998,
				Data:  "bizkit",
			},
			want: "ce0762697a6b697400",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := NewMVD2Writer(&pb.MvdDemo{})
			msg := writer.MarshalConfigstring(tc.data)
			got := hex.EncodeToString(msg.Data)
			if got != tc.want {
				t.Errorf("MarshalConfigstring(%v) = %s, want %s\n", tc.data, got, tc.want)
			}
		})
	}
}

func TestMvdMarshalConfigStrings(t *testing.T) {
	tests := []struct {
		name  string
		remap *pb.MvdConfigStringRemap
		data  map[int32]*pb.ConfigString
		want  string
	}{
		{
			name:  "two strings standard remap",
			remap: csRemap,
			data: map[int32]*pb.ConfigString{
				1800: {
					Index: 1800,
					Data:  "bonergarage",
				},
				1998: {
					Index: 1998,
					Data:  "bizkit",
				},
			},
			want: "0807626f6e657267617261676500ce0762697a6b6974002008",
		},
		{
			name:  "two strings extended remap",
			remap: csRemapNew,
			data: map[int32]*pb.ConfigString{
				1800: {
					Index: 1800,
					Data:  "bonergarage",
				},
				1998: {
					Index: 1998,
					Data:  "bizkit",
				},
			},
			want: "0807626f6e657267617261676500ce0762697a6b6974003e35",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := NewMVD2Writer(&pb.MvdDemo{
				Remap: tc.remap,
			})
			msg := writer.MarshalConfigstrings(tc.data)
			got := hex.EncodeToString(msg.Data)
			if got != tc.want {
				t.Errorf("MarshalConfigstrings(%v) = %s, want %s\n", tc.data, got, tc.want)
			}
		})
	}
}
