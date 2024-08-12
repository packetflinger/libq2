package demo

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/packetflinger/libq2/message"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/packetflinger/libq2/proto"
)

func TestMarshal(t *testing.T) {
	demo, e := OpenDM2File("../testdata/testduel.dm2")
	if e != nil {
		t.Errorf("%v", e)
	}
	textpb := &pb.TestDemo{}
	for {
		lump, size, err := nextLump(demo.Handle, int64(demo.Position))
		if err != nil {
			t.Error(err)
		}
		if size == 0 {
			break
		}

		demo.Position += size
		err = demo.Marshal(textpb, message.NewMessageBuffer(lump))
		if err != nil {
			t.Error(err)
		}
	}

	if demo.Handle == nil {
		t.Error("Handle - file handle is nil")
	}

	demo.Close()
}

func TestServerDataToProto(t *testing.T) {
	tests := []struct {
		name string
		in   *message.MessageBuffer
		want *pb.ServerInfo
	}{
		{
			name: "test 1",
			in: &message.MessageBuffer{
				Buffer: []byte{
					34, 0, 0, 0, 66, 133, 164, 102,
					1, 0, 0, 0, 84, 104, 101, 32,
					69, 100, 103, 101, 0,
				},
				Length: 21,
			},
			want: &pb.ServerInfo{
				Protocol:     34,
				ServerCount:  1722058050,
				Demo:         true,
				GameDir:      "",
				ClientNumber: 0,
				MapName:      "The Edge",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ServerDataToProto(tc.in)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestServerDataToBinary(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.ServerInfo
		want message.MessageBuffer
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
			want: message.MessageBuffer{
				Buffer: []byte{
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
			if !bytes.Equal(got.Buffer, tc.want.Buffer) {
				t.Error("\ngot:\n", got, "\nwant\n", tc.want)
			}
		})
	}
}

func TestParseEntityBitmask(t *testing.T) {
	tests := []struct {
		name string
		in   *message.MessageBuffer
		want uint32
	}{
		{
			name: "test 1",
			in: &message.MessageBuffer{
				Buffer: []byte{
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

func TestDeltaEntityBitmask(t *testing.T) {
	tests := []struct {
		name string
		to   *pb.PackedEntity
		from *pb.PackedEntity
		want uint32
	}{
		{
			name: "test 1",
			to: &pb.PackedEntity{
				OriginX: 3,
				AngleY:  34,
				Event:   4,
			},
			from: &pb.PackedEntity{
				OriginX: 4,
				AngleY:  35,
				Event:   5,
			},
			want: message.EntityOrigin1 | message.EntityAngle2 | message.EntityEvent,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DeltaEntityBitmask(tc.to, tc.from)
			if got != tc.want {
				t.Error("\ngot:\n", got, "\nwant\n", tc.want)
			}
		})
	}
}

func TestEntityToProto(t *testing.T) {
	tests := []struct {
		name string
		in   *message.MessageBuffer
		want *pb.PackedEntity
	}{
		{
			name: "test 1",
			in: &message.MessageBuffer{
				Buffer: []byte{
					131, 130, 128, 4, 22, 160, 7, 160, 243, 96, 16, 43,
				},
				Length: 12,
			},
			want: &pb.PackedEntity{
				Number:  22,
				OriginX: 1952,
				OriginY: -3168,
				OriginZ: 4192,
				Sound:   43,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bits := tc.in.ParseEntityBitmask()
			num := tc.in.ParseEntityNumber(bits)
			got := EntityToProto(tc.in, bits, num)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestDeltaPlayerstateBitmask(t *testing.T) {
	tests := []struct {
		name string
		to   *pb.PackedPlayer
		from *pb.PackedPlayer
		want uint16
	}{
		{
			name: "test 1",
			to: &pb.PackedPlayer{
				ViewAnglesX: 5,
				ViewOffsetX: 19,
			},
			from: &pb.PackedPlayer{
				ViewAnglesX: 2,
				ViewOffsetX: 25,
			},
			want: message.PlayerViewAngles | message.PlayerViewOffset,
		},
		{
			name: "test 2",
			to: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					Gravity: 800,
				},
			},
			from: &pb.PackedPlayer{
				Fov: 102,
				Stats: []*pb.PlayerStat{
					{
						Index: 2,
						Value: 100,
					},
				},
			},
			want: message.PlayerGravity | message.PlayerFOV,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DeltaPlayerBitmask(tc.to, tc.from)
			if got != tc.want {
				t.Error("\ngot:\n", got, "\nwant\n", tc.want)
			}
		})
	}
}

func TestPlayerstateToProto(t *testing.T) {
	tests := []struct {
		name string
		in   *message.MessageBuffer
		want *pb.PackedPlayer
	}{
		{
			name: "test 1",
			in: &message.MessageBuffer{
				Buffer: []byte{
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
				Stats: []*pb.PlayerStat{
					{Index: 0, Value: 2},
					{Index: 1, Value: 100},
					{Index: 2},
					{Index: 3},
					{Index: 4},
					{Index: 5},
					{Index: 6, Value: 5},
					{Index: 7},
					{Index: 8},
					{Index: 9},
					{Index: 10},
					{Index: 11, Value: 5},
					{Index: 12, Value: 7},
					{Index: 13},
					{Index: 14},
					{Index: 15},
					{Index: 16},
					{Index: 17},
					{Index: 18},
					{Index: 19},
					{Index: 20},
					{Index: 21},
					{Index: 22},
					{Index: 23},
					{Index: 24},
					{Index: 25},
					{Index: 26},
					{Index: 27},
					{Index: 28},
					{Index: 29},
					{Index: 30},
					{Index: 31},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PlayerstateToProto(tc.in)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestPlayerstateToBinary(t *testing.T) {
	tests := []struct {
		name string
		want *message.MessageBuffer
		in   *pb.PackedPlayer
	}{
		{
			name: "test 1",
			want: &message.MessageBuffer{
				Buffer: []byte{
					17,    // svc_playerstate
					50, 8, // bitmask
					50, 0, 0, 0, 25, 0, // origin
					4,     // flags
					32, 3, // gravity
					105,        // fov
					0, 0, 0, 0, // stat bitmask (none)
				},
				Length: 47,
			},
			in: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 50,
					OriginY: 0,
					OriginZ: 25,
					Flags:   4,
					Gravity: 800,
				},
				Fov: 105,
			},
		},
		{
			name: "test 2 with stats",
			want: &message.MessageBuffer{
				Buffer: []byte{
					17,    // svc_playerstate
					32, 0, // bitmask
					32, 3, // gravity
					12, 0, 0, 0, // stat bitmask
					89, 0, 5, 0, // stats
				},
				Length: 13,
			},
			in: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					Gravity: 800,
				},
				Stats: []*pb.PlayerStat{
					{Index: 2, Value: 89},
					{Index: 3, Value: 5},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := message.MessageBuffer{}
			DeltaPlayer(&pb.PackedPlayer{}, tc.in, &got)
			if !bytes.Equal(got.Buffer, tc.want.Buffer) {
				t.Error("\ngot:\n", got, "\nwant\n", tc.want)
			}
		})
	}
}
