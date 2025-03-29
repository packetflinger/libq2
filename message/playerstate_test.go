package message

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/packetflinger/libq2/proto"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestDeltaPlayerBitmask(t *testing.T) {
	tests := []struct {
		name string
		to   *pb.PackedPlayer
		from *pb.PackedPlayer
		want int
	}{
		{
			name: "nil playerstates",
			to:   nil,
			from: nil,
			want: 0,
		},
		{
			name: "nil from",
			to: &pb.PackedPlayer{
				Fov: 100,
			},
			from: nil,
			want: 2048,
		},
		{
			name: "nil to",
			to:   nil,
			from: &pb.PackedPlayer{
				Fov: 100,
			},
			want: 2048,
		},
		{
			name: "gravity and fov different",
			to: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					Gravity: 800,
				},
				Fov:     105,
				RdFlags: 5,
			},
			from: &pb.PackedPlayer{
				Fov:     100,
				RdFlags: 5,
			},
			want: 2048 + 32,
		},
	}
	for _, tc := range tests {
		got := DeltaPlayerBitmask(tc.to, tc.from)
		if got != tc.want {
			t.Errorf("DeltaPlayerBitmask(%v, %v) = %v, want %v\n", prototext.Format(tc.to), prototext.Format(tc.from), got, tc.want)
		}
	}
}

func TestParseDeltaPlayerstate(t *testing.T) {
	tests := []struct {
		name string
		data string
		from *pb.PackedPlayer
		want *pb.PackedPlayer
	}{
		{
			name: "no data and nil from",
			data: "",
			from: nil,
			want: nil,
		},
		{
			name: "lots of data, no from",
			data: "96265F225013C10E5D07F10200000400003654000016FB0AF501FDFFFF000033028000005E000100",
			from: nil,
			want: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX:   8799,
					OriginY:   4944,
					OriginZ:   3777,
					VelocityX: 1885,
					VelocityY: 753,
					Flags:     4,
				},
				ViewOffsetZ: 54,
				KickAnglesX: 84,
				GunAnglesX:  1,
				GunAnglesY:  -3,
				GunAnglesZ:  -1,
				GunOffsetX:  -5,
				GunOffsetY:  10,
				GunOffsetZ:  -11,
				GunFrame:    22,
				BlendW:      -1,
				BlendZ:      51,
				Stats: map[uint32]int32{
					1:  94,
					15: 1,
				},
			},
		},
		{
			name: "lots of data with from",
			data: "96265F225013C10E5D07F10200000400003654000016FB0AF501FDFFFF000033028000005E000100",
			from: &pb.PackedPlayer{
				ViewAnglesX: 5,
				ViewAnglesZ: 90,
			},
			want: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX:   8799,
					OriginY:   4944,
					OriginZ:   3777,
					VelocityX: 1885,
					VelocityY: 753,
					Flags:     4,
				},
				ViewAnglesX: 5,
				ViewAnglesZ: 90,
				ViewOffsetZ: 54,
				KickAnglesX: 84,
				GunAnglesX:  1,
				GunAnglesY:  -3,
				GunAnglesZ:  -1,
				GunOffsetX:  -5,
				GunOffsetY:  10,
				GunOffsetZ:  -11,
				GunFrame:    22,
				BlendW:      -1,
				BlendZ:      51,
				Stats: map[uint32]int32{
					1:  94,
					15: 1,
				},
			},
		},
		{
			name: "another",
			data: "11F2393B239F13C10E0420030000006000000000587D038C0F0000041EFB0AF50000006900000000",
			want: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					Type:  57,
					Flags: 59,
				},
				KickAnglesX: 35,
				KickAnglesY: -97,
				KickAnglesZ: 19,
				GunOffsetX:  4,
				GunOffsetY:  32,
				GunOffsetZ:  3,
				GunIndex:    193,
				GunFrame:    14,
				RdFlags:     96,
			},
		},
	}
	for _, tc := range tests {
		rawbytes, err := hex.DecodeString(tc.data)
		if err != nil {
			t.Errorf("TestParseDeltaPlayerstate() - error decoding input string: %v\n", err)
		}
		buf := NewBuffer(rawbytes)
		got := buf.ParseDeltaPlayerstate(tc.from)
		if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
			t.Errorf("(%v).ParseDeltaPlayerstate(%v) = %v, want %v\n", buf, prototext.Format(tc.from), got, prototext.Format(tc.want))
		}
	}
}

func TestWriteDeltaPlayerstate(t *testing.T) {
	tests := []struct {
		name string
		to   *pb.PackedPlayer
		from *pb.PackedPlayer
		want string
	}{
		{
			name: "nil playerstates",
			to:   nil,
			from: nil,
			want: "11000000000000",
		},
		{
			name: "small diff no stats",
			to: &pb.PackedPlayer{
				RdFlags: 96,
			},
			from: &pb.PackedPlayer{
				RdFlags: 99,
			},
			want: "1100406000000000",
		},
		{
			name: "small diff with stats",
			to: &pb.PackedPlayer{
				RdFlags: 96,
				Stats: map[uint32]int32{
					1: 35,
					2: 9,
				},
			},
			from: &pb.PackedPlayer{
				RdFlags: 99,
				Stats: map[uint32]int32{
					1: 35,
					2: 5,
				},
			},
			want: "11004060040000000900",
		},
	}
	for _, tc := range tests {
		got := hex.EncodeToString(WriteDeltaPlayerstate(tc.from, tc.to).Data)
		if got != tc.want {
			t.Errorf("WriteDeltaPlayerstate(%v, %v) = %v, want %v\n", prototext.Format(tc.from), prototext.Format(tc.to), got, tc.want)
		}
	}
}
