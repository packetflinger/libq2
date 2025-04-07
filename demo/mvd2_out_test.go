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
		want string
	}{
		{
			name: "no demo flags",
			demo: &pb.MvdDemo{
				Version:  2010,
				Identity: 123456789,
				GameDir:  "uranus",
				Dummy:    20,
			},
			want: "25000000da0715cd5b077572616e7573001400",
		},
		{
			name: "with demo flags",
			demo: &pb.MvdDemo{
				Version:  2012,
				Identity: 123456789,
				GameDir:  "uranus",
				Dummy:    20,
				Flags:    715,
			},
			want: "25000000dc07cb0215cd5b077572616e7573001400",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := NewMVD2Writer(tc.demo)
			msg := writer.MarshalServerData()
			got := hex.EncodeToString(msg.Data)
			if got != tc.want {
				t.Errorf("MarshalServerData(%v) = %s, want %s\n", tc.demo, got, tc.want)
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

func TestMvdMarshalPlayer(t *testing.T) {
	tests := []struct {
		name   string
		number int32
		from   *pb.PackedPlayer
		to     *pb.PackedPlayer
		want   string
	}{
		{
			name:   "player2",
			number: 2,
			from: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 5,
					OriginY: 25,
				},
				Fov: 105,
				Stats: map[uint32]int32{
					1: 95,
				},
			},
			to: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 10,
					OriginY: 25,
				},
				Fov: 105,
				Stats: map[uint32]int32{
					1: 100,
				},
			},
			want: "0202001102000a0019000000020000006400",
		},
		{
			name:   "player3",
			number: 3,
			from: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 200,
					OriginY: -4,
				},
				Fov: 110,
				Stats: map[uint32]int32{
					3: 50,
				},
			},
			to: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 10,
					OriginY: 25,
				},
				Fov: 105,
				Stats: map[uint32]int32{
					1: 100,
				},
			},
			want: "0302081102080a0019000000690a00000064000000",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := NewMVD2Writer(&pb.MvdDemo{
				Players: map[int32]*pb.MvdPlayer{
					tc.number: {
						Name:        "claire",
						PlayerState: tc.from,
					},
				},
			})
			msg := writer.MarshalPlayer(tc.number, tc.to)
			got := hex.EncodeToString(msg.Data)
			if got != tc.want {
				t.Errorf("MarshalPlayer(%v, %v) = %s, want %s\n", tc.number, tc.to, got, tc.want)
			}
		})
	}
}

func TestMvdMarshalPlayers(t *testing.T) {
	tests := []struct {
		name    string
		from    map[int32]*pb.MvdPlayer
		players map[int32]*pb.PackedPlayer
		want    string
	}{
		{
			name: "player2",
			from: map[int32]*pb.MvdPlayer{
				3: {
					Name: "claire",
					PlayerState: &pb.PackedPlayer{
						Movestate: &pb.PlayerMove{
							OriginX: 5,
							OriginY: 25,
						},
						Fov: 105,
						Stats: map[uint32]int32{
							1: 95,
						},
					},
				},
				4: {
					Name: "someone_else",
					PlayerState: &pb.PackedPlayer{
						Movestate: &pb.PlayerMove{
							OriginX: 10,
							OriginY: 25,
						},
						Fov: 105,
						Stats: map[uint32]int32{
							1: 100,
						},
					},
				},
			},
			players: map[int32]*pb.PackedPlayer{
				3: {
					Movestate: &pb.PlayerMove{
						OriginX: 5,
						OriginY: 25,
					},
					Fov: 105,
					Stats: map[uint32]int32{
						1: 95,
					},
				},
				4: {
					Movestate: &pb.PlayerMove{
						OriginX: 10,
						OriginY: 25,
					},
					Fov: 105,
					Stats: map[uint32]int32{
						1: 100,
					},
				},
			},
			want: "????",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := NewMVD2Writer(&pb.MvdDemo{
				Players: tc.from,
			})
			msg := writer.MarshalPlayers(tc.players)
			got := hex.EncodeToString(msg.Data)
			if got != tc.want {
				t.Errorf("MarshalPlayers(%v) = %s, want %s\n", tc.players, got, tc.want)
			}
		})
	}
}
