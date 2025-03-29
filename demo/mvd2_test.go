package demo

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/packetflinger/libq2/message"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/packetflinger/libq2/proto"
)

/*
func TestMVDUnmarshal(t *testing.T) {
	parser, err := NewMVD2Demo("../testdata/test.mvd2")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = parser.Unmarshal()
	if err != nil {
		t.Error(err)
	}
}
*/

func TestMvdParseConfigstrings(t *testing.T) {
	tests := []struct {
		name string
		data string
		want map[uint32]*pb.ConfigString
	}{
		{
			name: "2_strings_complete_proper_ending",
			data: "000061626300010078797a002008",
			want: map[uint32]*pb.ConfigString{
				0: {Index: 0, Data: "abc"},
				1: {Index: 1, Data: "xyz"},
			},
		},
		{
			name: "2_strings_invalid_ending",
			data: "000061626300010078797a00",
			want: map[uint32]*pb.ConfigString{
				0: {Index: 0, Data: "abc"},
				1: {Index: 1, Data: "xyz"},
			},
		},
		{
			name: "maxclients_and_players",
			data: "1E003235003F04416D6D6F205061636B0040044865616C746800200557616C6C466C795B425A5A5A5D5C6D616C652F6772756E740021055B647265616D5D73686C6F6F5C6D616C652F6772756E74002205636C616972655C66656D616C652F617468656E610034055B4D5644535045435D5C6D616C652F6772756E74002008",
			want: map[uint32]*pb.ConfigString{
				30:   {Index: 30, Data: "25"},
				1087: {Index: 1087, Data: "Ammo Pack"},
				1088: {Index: 1088, Data: "Health"},
				1312: {Index: 1312, Data: "Wallfly[BZZZ]\\male/grunt"},
				1313: {Index: 1313, Data: "[dream]shloo\\male/grunt"},
				1314: {Index: 1314, Data: "claire\\female/athena"},
				1332: {Index: 1332, Data: "[MVDSPEC]\\male/grunt"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Error("error decoding hex string")
			}
			buf := message.NewBuffer(in)
			parser := MVD2Parser{
				demo: &pb.MvdDemo{
					Remap: csRemap,
				},
			}
			got := parser.ParseConfigStrings(&buf)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("ParseConfigStrings(%v) = \n%v,\nwant \n%v\n", &buf, got, tc.want)
			}
		})
	}
}

func TestMvdParseEntityBits(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		parser   *MVD2Parser
		wantBits uint64
		wantNum  uint32
	}{
		{
			name: "empty",
			data: "0000",
			parser: &MVD2Parser{
				demo: &pb.MvdDemo{
					EntityStateFlags: 0,
				},
			},
			wantBits: 0,
			wantNum:  0,
		},
		{
			name: "1 byte mask",
			data: "0c02",
			parser: &MVD2Parser{
				demo: &pb.MvdDemo{
					EntityStateFlags: 0,
				},
			},
			wantBits: message.EntityAngle2 | message.EntityAngle3,
			wantNum:  2,
		},
		{
			name: "2 byte mask",
			data: "8c0202",
			parser: &MVD2Parser{
				demo: &pb.MvdDemo{
					EntityStateFlags: 0,
				},
			},
			wantBits: message.EntityAngle2 | message.EntityAngle3 | message.EntityMoreBits1 | message.EntityOrigin3,
			wantNum:  2,
		},
		{
			name: "3 byte mask",
			data: "80801002",
			parser: &MVD2Parser{
				demo: &pb.MvdDemo{
					EntityStateFlags: 0,
				},
			},
			wantBits: message.EntityMoreBits1 | message.EntityMoreBits2 | message.EntityModel2,
			wantNum:  2,
		},
		{
			name: "4 byte mask",
			data: "8080800202",
			parser: &MVD2Parser{
				demo: &pb.MvdDemo{
					EntityStateFlags: 0,
				},
			},
			wantBits: message.EntityMoreBits1 | message.EntityMoreBits2 | message.EntityMoreBits3 | message.EntitySkin16,
			wantNum:  2,
		},
		{
			name: "5 byte mask",
			data: "808080800102",
			parser: &MVD2Parser{
				demo: &pb.MvdDemo{
					EntityStateFlags: 0,
				},
			},
			wantBits: message.EntityMoreBits1 | message.EntityMoreBits2 | message.EntityMoreBits3 | message.EntityMoreBits4 | message.EntityScale,
			wantNum:  2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Error("error decoding hex string")
			}
			buf := message.NewBuffer(in)
			gotNum, gotBits := tc.parser.ParseEntityBits(&buf)
			if gotBits != tc.wantBits || gotNum != tc.wantNum {
				t.Errorf("ParseEntityBits(%v) = (%v, %v) want (%v, %v)\n", &buf, gotNum, gotBits, tc.wantNum, tc.wantBits)
			}
		})
	}
}

func TestMvdParsePacketPlayersFromSkins(t *testing.T) {
	tests := []struct {
		name string
		data string
		want map[uint32]*pb.ConfigString
	}{
		{
			name: "2_strings_complete_proper_ending",
			data: "000061626300010078797a002008",
			want: map[uint32]*pb.ConfigString{
				0: {Index: 0, Data: "abc"},
				1: {Index: 1, Data: "xyz"},
			},
		},
		{
			name: "2_strings_invalid_ending",
			data: "000061626300010078797a00",
			want: map[uint32]*pb.ConfigString{
				0: {Index: 0, Data: "abc"},
				1: {Index: 1, Data: "xyz"},
			},
		},
		{
			name: "maxclients_and_players",
			data: "1E003235003F04416D6D6F205061636B0040044865616C746800200557616C6C466C795B425A5A5A5D5C6D616C652F6772756E740021055B647265616D5D73686C6F6F5C6D616C652F6772756E74002205636C616972655C66656D616C652F617468656E610034055B4D5644535045435D5C6D616C652F6772756E74002008",
			want: map[uint32]*pb.ConfigString{
				30:   {Index: 30, Data: "25"},
				1087: {Index: 1087, Data: "Ammo Pack"},
				1088: {Index: 1088, Data: "Health"},
				1312: {Index: 1312, Data: "Wallfly[BZZZ]\\male/grunt"},
				1313: {Index: 1313, Data: "[dream]shloo\\male/grunt"},
				1314: {Index: 1314, Data: "claire\\female/athena"},
				1332: {Index: 1332, Data: "[MVDSPEC]\\male/grunt"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Error("error decoding hex string")
			}
			buf := message.NewBuffer(in)
			parser := MVD2Parser{
				demo: &pb.MvdDemo{
					Remap: csRemap,
				},
			}
			got := parser.ParseConfigStrings(&buf)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("ParseConfigStrings(%v) = \n%v,\nwant \n%v\n", &buf, got, tc.want)
			}
		})
	}
}

func TestMvdParsePacketPlayers(t *testing.T) {
	tests := []struct {
		name string
		data string
		want map[uint32]*pb.PackedPlayer
	}{
		{
			name: "test0",
			data: "011E577B21812AEE1B0000584B0950BC3B2E0003016E7F1800840200640007002F001B0091000A0016000E0026062406021E47612D7F37C1180000588C072E8C2019697F387E840200C4000B0032001B0064000A000A000E0001006400C8003200C8003200320026062406141E5100270015C80E0000580000004000DCEE5A012000840200010026062406FF",
			want: map[uint32]*pb.PackedPlayer{
				0: {Fov: 90},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Error("error decoding hex string")
			}
			buf := message.NewBuffer(in)
			parser := MVD2Parser{
				demo: &pb.MvdDemo{
					Remap: csRemap,
					Players: map[uint32]*pb.MvdPlayer{
						0: {Name: "Wallfly[BZZZ]"},
						1: {Name: "[dream]shloo"},
						2: {Name: "claire"},
						3: {Name: "[MVDSPEC]"},
					},
				},
			}
			got, err := parser.ParsePacketPlayers(&buf)
			if err != nil {
				t.Errorf("error: %v\n", err)
			}
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("ParsePacketPlayers(%v) = \n%v,\nwant \n%v\n", &buf, got, tc.want)
			}
		})
	}
}

func TestMvdUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		demofile string
	}{
		{
			name:     "test1",
			demofile: "../testdata/test.mvd2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewMVD2Parser(tc.demofile)
			if err != nil {
				t.Errorf("error creating parser: %v", err)
			}
			demo, err := parser.Unmarshal()
			if err != nil {
				t.Errorf("error unmarshalling: %v", err)
			}
			if demo == nil {
				t.Error()
			}
			//fmt.Println(prototext.Format(demo))
			//t.Error()
		})
	}
}
