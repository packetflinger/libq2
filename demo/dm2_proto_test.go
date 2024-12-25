package demo

import (
	"bytes"
	"testing"

	"github.com/packetflinger/libq2/message"

	pb "github.com/packetflinger/libq2/proto"
)

func TestServerDataToBinary(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.ServerInfo
		want message.Buffer
	}{
		{
			name: "test 1",
			in: &pb.ServerInfo{
				Protocol:     34,
				ServerCount:  1722058050,
				Demo:         true,
				GameDir:      "",
				ClientNumber: 0,
				MapName:      "The Edge",
			},
			want: message.Buffer{
				Data: []byte{
					12, 34, 0, 0, 0, 66, 133, 164, 102,
					1, 0, 0, 0, 84, 104, 101, 32,
					69, 100, 103, 101, 0,
				},
				Length: 22,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ServerDataToBinary(tc.in)
			if !bytes.Equal(got.Data, tc.want.Data) {
				t.Error("\ngot:\n", got, "\nwant\n", tc.want)
			}
		})
	}
}

func TestParseEntityBitmask(t *testing.T) {
	tests := []struct {
		name string
		in   *message.Buffer
		want uint32
	}{
		{
			name: "test 1",
			in: &message.Buffer{
				Data: []byte{
					131, 130, 128, 4,
				},
				Length: 4,
			},
			want: message.EntityOrigin1 |
				message.EntityOrigin2 |
				message.EntityOrigin3 |
				message.EntitySound |
				message.EntityMoreBits1 |
				message.EntityMoreBits2 |
				message.EntityMoreBits3,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.ParseEntityBitmask()
			if got != tc.want {
				t.Error("\ngot:\n", got, "\nwant\n", tc.want)
			}
		})
	}
}

func TestPlayerstateToProto(t *testing.T) {
	tests := []struct {
		name string
		in   *message.Buffer
		want *pb.PackedPlayer
	}{
		{
			name: "test 1",
			in: &message.Buffer{
				Data: []byte{
					242, 57, 169, 48, 143, 31, 193, 32, // bitmask, origin
					4, 32, 3, 0, 0, 0, 128, 0, 0, // flags, gravity, delta angles
					0, 0, 88, 168, 13, 157, 181, 0, 0, // view offset, view angles
					4, 48, 12, 0, 245, 0, 0, 0, // gun index, gun frame, gun offset, gun angles
					105, 67, 24, 0, 0, // fov, stat bitmask (6211 == 0, 1, 6, 11, 12),
					2, 0, 100, 0, 5, 0, 5, 0, 7, 0, // stats
				},
				Length: 47,
			},
			want: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX:     12457,
					OriginY:     8079,
					OriginZ:     8385,
					Flags:       4,
					Gravity:     800,
					DeltaAngleX: 0,
					DeltaAngleY: -32768,
					DeltaAngleZ: 0,
				},
				ViewOffsetX: 0,
				ViewOffsetY: 0,
				ViewOffsetZ: 88,
				ViewAnglesX: 3496,
				ViewAnglesY: -19043,
				ViewAnglesZ: 0,
				GunIndex:    4,
				GunFrame:    48,
				GunOffsetX:  12,
				GunOffsetY:  0,
				GunOffsetZ:  -11,
				GunAnglesX:  0,
				GunAnglesY:  0,
				GunAnglesZ:  0,
				Fov:         105,
				Stats: map[uint32]int32{
					0:  2,
					1:  100,
					6:  5,
					11: 5,
					12: 7,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			//ps := tc.in.ParseDeltaPlayerstate()
			//got := PlayerstateToProto(tc.in)
			//if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
			//	t.Error(diff)
			//}
		})
	}
}
