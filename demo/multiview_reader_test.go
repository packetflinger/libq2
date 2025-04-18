package demo

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/packetflinger/libq2/message"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/packetflinger/libq2/proto"
)

func TestMvdParseServerData(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.MvdServerData
	}{
		{
			name: "test0",
			data: "2425000000DA07B5218E5E6F70656E74646D00140000",
			want: &pb.MvdServerData{
				Protocol:         2010,
				Identity:         1586373045,
				GameDirectory:    "opentdm",
				DummyClient:      20,
				Flags:            1,
				EntitystateFlags: 48,
				Remap:            csRemap,
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
				demo: &pb.MvdDemo{},
			}
			got, err := parser.ParseServerData(&buf, (buf.ReadByte() >> CommandBits))
			if err != nil {
				t.Errorf("ParseServerData(%v) error: %v", tc.data, err)
			}
			// parser.demo.Remap = nil // don't care about cs map
			if diff := cmp.Diff(got, tc.want, protocmp.Transform(), protocmp.IgnoreFields(parser.demo.Remap)); diff != "" {
				t.Errorf("ParseServerData(%v) = \n%v,\nwant \n%v\n", &buf, got, tc.want)
			}
		})
	}
}

func TestMvdParseConfigstrings(t *testing.T) {
	tests := []struct {
		name string
		data string
		want map[int32]*pb.ConfigString
	}{
		{
			name: "2_strings_complete_proper_ending",
			data: "000061626300010078797a002008",
			want: map[int32]*pb.ConfigString{
				0: {Index: 0, Data: "abc"},
				1: {Index: 1, Data: "xyz"},
			},
		},
		{
			name: "2_strings_invalid_ending",
			data: "000061626300010078797a00",
			want: map[int32]*pb.ConfigString{
				0: {Index: 0, Data: "abc"},
				1: {Index: 1, Data: "xyz"},
			},
		},
		{
			name: "maxclients_and_players",
			data: "1E003235003F04416D6D6F205061636B0040044865616C746800200557616C6C466C795B425A5A5A5D5C6D616C652F6772756E740021055B647265616D5D73686C6F6F5C6D616C652F6772756E74002205636C616972655C66656D616C652F617468656E610034055B4D5644535045435D5C6D616C652F6772756E74002008",
			want: map[int32]*pb.ConfigString{
				30:   {Index: 30, Data: "25"},
				1087: {Index: 1087, Data: "Ammo Pack"},
				1088: {Index: 1088, Data: "Health"},
				1312: {Index: 1312, Data: "WallFly[BZZZ]\\male/grunt"},
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
			got := parser.ParseConfigStrings(&buf, csRemap)
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
		wantBits int64
		wantNum  int32
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
					EntityStateFlags: 128,
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
		want map[int32]*pb.ConfigString
	}{
		{
			name: "2_strings_complete_proper_ending",
			data: "010061626300020078797a002008",
			want: map[int32]*pb.ConfigString{
				1: {Index: 1, Data: "abc"},
				2: {Index: 2, Data: "xyz"},
			},
		},
		{
			name: "2_strings_invalid_ending",
			data: "000061626300010078797a00",
			want: map[int32]*pb.ConfigString{
				0: {Index: 0, Data: "abc"},
				1: {Index: 1, Data: "xyz"},
			},
		},
		{
			name: "maxclients_and_players",
			data: "1E003235003F04416D6D6F205061636B0040044865616C746800200557616C6C466C795B425A5A5A5D5C6D616C652F6772756E740021055B647265616D5D73686C6F6F5C6D616C652F6772756E74002205636C616972655C66656D616C652F617468656E610034055B4D5644535045435D5C6D616C652F6772756E74002008",
			want: map[int32]*pb.ConfigString{
				30:   {Index: 30, Data: "25"},
				1087: {Index: 1087, Data: "Ammo Pack"},
				1088: {Index: 1088, Data: "Health"},
				1312: {Index: 1312, Data: "WallFly[BZZZ]\\male/grunt"},
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
			got := parser.ParseConfigStrings(&buf, csRemap)
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
		want map[int32]*pb.PackedPlayer
	}{
		{
			name: "test0",
			data: "011E577B21812AEE1B0000584B0950BC3B2E0003016E7F1800840200640007002F001B0091000A0016000E0026062406021E47612D7F37C1180000588C072E8C2019697F387E840200C4000B0032001B0064000A000A000E0001006400C8003200C8003200320026062406141E5100270015C80E0000580000004000DCEE5A012000840200010026062406FF",
			want: map[int32]*pb.PackedPlayer{
				1: {
					Movestate: &pb.PlayerMove{
						OriginX: 8571,
						OriginY: 10881,
						OriginZ: 7150,
					},
					ViewAnglesX: 2379,
					ViewAnglesY: -17328,
					ViewOffsetZ: 88,
					GunAnglesY:  3,
					GunAnglesZ:  1,
					GunIndex:    59,
					GunFrame:    46,
					Fov:         110,
					Stats: map[uint32]int32{
						0:  2,
						1:  100,
						2:  7,
						3:  47,
						4:  27,
						5:  145,
						6:  10,
						11: 22,
						12: 14,
						26: 1574,
						31: 1572,
					},
				},
				2: {
					Movestate: &pb.PlayerMove{
						OriginX: 11617,
						OriginY: 14207,
						OriginZ: 6337,
					},
					ViewAnglesX: 1932,
					ViewAnglesY: -29650,
					ViewOffsetZ: 88,
					GunIndex:    32,
					GunFrame:    25,
					Fov:         105,
					Stats: map[uint32]int32{
						0:  2,
						1:  196,
						2:  11,
						3:  50,
						4:  27,
						5:  100,
						6:  10,
						11: 10,
						12: 14,
						13: 1,
						17: 100,
						18: 200,
						19: 50,
						20: 200,
						21: 50,
						22: 50,
						26: 1574,
						31: 1572,
					},
				},
				20: {
					Movestate: &pb.PlayerMove{
						OriginX: 9984,
						OriginY: 5376,
						OriginZ: 3784,
					},
					ViewAnglesY: 16384,
					ViewOffsetZ: 88,
					GunAnglesY:  -36,
					GunAnglesZ:  -18,
					Fov:         90,
					Stats: map[uint32]int32{
						0:  2,
						13: 1,
						26: 1574,
						31: 1572,
					},
				},
			},
		},
		{
			name: "leading zero",
			data: "001ED100270015C80E0000580000004000DCEE5A012000840200010026062406011E478030C012410B0000580000006004276E431800840200640005000500070026062406021ED100270015C80E0000580000004000DCEE78012000840200010026062406035E57382D01FD0E1D0000607F0596730700FA3B1D08F5FB787F180084020064000700260009002F000A0016000E0026062406041E474A136D2AC1180000587E0330B1201C6F7F38008C020064000B0032001B00640016000A0010000100260627062406051EC746290125011D000058000000202C0D697F38008402005C000F00C8001B003C00100010000A00010026062406065E57CD0FFB1EC11800005DBEF7D2440400FE201905F5FB787F187E84020064000B0032001B0064000E000A0008006400C8003200C8003200320026062406071E5721358A1E790F00005D6A08DCC03221070C06697F18008402006200080019001B002D0012001300090026062406081EC7003B001701110000580000008030056E7F38008402006400060064001B006400120012000900010026062406091ED100270015C80E0000580000004000DCEE6E0120008402000100260624060A5E57DC28F330C11A00005DAF092B74030002201905F6FB647F180084020064000B0031001B005B000D000A000F00260624060B1ED100270015C80E0000580000004000DCEE640120008402000100260624060C5ED7F031BCF8C11400005B58FF9F7F0300FF201903FC FE5A7F10 00940200 64000B00 2E001B00 64000A00 0E002606 28062406 0D1ED100 270015C8 0E000058 00000040 00DCEE5D 01200084 02000100 26062406 0E5ED7CA 276223C1 1400005D B0012B01 0400022C 17050B05 5A7F1000 84020064 000F00C8 001B0089 000A000E 00260624 06141F41 01002700 15C80E00 00580000 00405A01 20008402 00010026 062406FF",
			want: map[int32]*pb.PackedPlayer{
				1: {
					Movestate: &pb.PlayerMove{
						OriginX: 8571,
						OriginY: 10881,
						OriginZ: 7150,
					},
					ViewAnglesX: 2379,
					ViewAnglesY: -17328,
					ViewOffsetZ: 88,
					GunAnglesY:  3,
					GunAnglesZ:  1,
					GunIndex:    59,
					GunFrame:    46,
					Fov:         110,
					Stats: map[uint32]int32{
						0:  2,
						1:  100,
						2:  7,
						3:  47,
						4:  27,
						5:  145,
						6:  10,
						11: 22,
						12: 14,
						26: 1574,
						31: 1572,
					},
				},
				2: {
					Movestate: &pb.PlayerMove{
						OriginX: 11617,
						OriginY: 14207,
						OriginZ: 6337,
					},
					ViewAnglesX: 1932,
					ViewAnglesY: -29650,
					ViewOffsetZ: 88,
					GunIndex:    32,
					GunFrame:    25,
					Fov:         105,
					Stats: map[uint32]int32{
						0:  2,
						1:  196,
						2:  11,
						3:  50,
						4:  27,
						5:  100,
						6:  10,
						11: 10,
						12: 14,
						13: 1,
						17: 100,
						18: 200,
						19: 50,
						20: 200,
						21: 50,
						22: 50,
						26: 1574,
						31: 1572,
					},
				},
				20: {
					Movestate: &pb.PlayerMove{
						OriginX: 9984,
						OriginY: 5376,
						OriginZ: 3784,
					},
					ViewAnglesY: 16384,
					ViewOffsetZ: 88,
					GunAnglesY:  -36,
					GunAnglesZ:  -18,
					Fov:         90,
					Stats: map[uint32]int32{
						0:  2,
						13: 1,
						26: 1574,
						31: 1572,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := strings.ReplaceAll(tc.data, " ", "")
			in, err := hex.DecodeString(data)
			if err != nil {
				t.Error("error decoding hex string")
			}
			buf := message.NewBuffer(in)
			parser := MVD2Parser{
				demo: &pb.MvdDemo{
					Remap: csRemap,
					Players: map[int32]*pb.MvdPlayer{
						0:  {Name: "Wallfly[BZZZ]"},
						1:  {Name: "[dream]shloo"},
						2:  {Name: "claire"},
						20: {Name: "[MVDSPEC]"},
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
		name      string
		demofile  string
		democount int
	}{
		{
			name:      "test1",
			demofile:  "../testdata/test.mvd2",
			democount: 1,
		},
		{
			name:      "test2",
			demofile:  "../testdata/big.mvd2",
			democount: 7,
		},
		{
			name:      "gzip test",
			demofile:  "../testdata/ziptest.mvd2.gz",
			democount: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewMVD2Parser(tc.demofile)
			if err != nil {
				t.Errorf("error creating parser: %v", err)
			}
			parser.debug = false
			demos, err := parser.Unmarshal()
			if err != nil {
				t.Errorf("error unmarshalling: %v", err)
			}
			if demos == nil {
				t.Error()
			}
			if len(demos) != tc.democount {
				t.Errorf("demo count mismatch, got %d, want %d", len(demos), tc.democount)
			}
		})
	}
}
