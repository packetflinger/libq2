package message

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/packetflinger/libq2/proto"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseChallenge(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    *pb.Challenge
		wantErr bool
	}{
		{
			name: "valid all three protocols",
			data: "ffffffff6368616c6c656e67652039313039303836343420703d33342c33352c3336",
			want: &pb.Challenge{
				Number:    910908644,
				Protocols: []int32{34, 35, 36},
			},
			wantErr: false,
		},
		{
			name: "valid only two protocols",
			data: "ffffffff6368616c6c656e67652032363832393035313320703d33342c3335",
			want: &pb.Challenge{
				Number:    268290513,
				Protocols: []int32{34, 35},
			},
			wantErr: false,
		},
		{
			name: "valid no protocols",
			data: "ffffffff6368616c6c656e676520323638323930353133",
			want: &pb.Challenge{
				Number:    268290513,
				Protocols: []int32{},
			},
			wantErr: false,
		},
		{
			name:    "invalid sequence",
			data:    "6368616c6c656e67652032363832393035313320703d33342c3335",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			msg := NewBuffer(bytes)
			got, err := msg.ParseChallenge()
			if err != nil && !tc.wantErr {
				t.Error(err)
			}
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("ParseChallenge() error (-got/+want):\n%v\n", diff)
			}
		})
	}
}

/*
func TestServerData(t *testing.T) {
	msg := Buffer{
		Data: []byte{
			0x22, 0x00, 0x00, 0x00, 0x42, 0x85, 0xA4, 0x66,
			0x01, 0x00, 0x00, 0x00, 0x54, 0x68, 0x65, 0x20,
			0x45, 0x64, 0x67, 0x65, 0x00},
	}

	// read it
	sd := msg.ParseServerData()
	if sd.Protocol != 34 {
		t.Error("parse error, wrong protocol. Got", sd.Protocol, "want 34")
	}
	if sd.ServerCount != 1722058050 {
		t.Error("parse error, wrong client number. Got", sd.ServerCount, "want 1722058050")
	}
	if sd.Demo != 1 {
		t.Error("parse error, not demo. Got", sd.Demo, "want 1")
	}
	if sd.GameDir != "" {
		t.Error("parse error, wrong gamedir. Got", sd.GameDir, "want \"\"")
	}
	if sd.ClientNumber != 0 {
		t.Error("parse error, wrong clientnum. Got", sd.ClientNumber, "want 0")
	}
	if sd.MapName != "The Edge" {
		t.Error("parse error, wrong mapname. Got", sd.MapName, "want The Edge")
	}
	msg.Reset()

	// write it
	got := sd.Marshal()
	got.Reset()
	if diff := cmp.Diff(&msg, got); diff != "" {
		t.Error("marshal error, got diff\n", diff)
	}
}
*/

/*
func TestParseConfigstring(t *testing.T) {
	msg := Buffer{
		Data: []byte{
			0x02, 0x00, 0x75, 0x6E, 0x69, 0x74, 0x31, 0x5F, 0x00},
	}
	cs := msg.ParseConfigString()
	if cs.Index != 2 {
		t.Error("parseconfigstring, wrong index, got", cs.Index, "want 2")
	}
	if cs.String != "unit1_" {
		t.Error("parseconfigstring, wrong string, got", cs.String, "want unit1_")
	}

	got := cs.Marshal()
	got.Reset()
	msg.Reset()

	if diff := cmp.Diff(&msg, got); diff != "" {
		t.Error("marshal error, got diff\n", diff)
	}
}
*/

/*
func TestParseBaseline(t *testing.T) {
	msg := Buffer{
		Data: []byte{0x83, 0xCA, 0x04, 0x23, 0x1C, 0x01, 0x00, 0x02, 0x80, 0x21, 0x00, 0x28, 0x78, 0x1C},
	}
	entity := msg.ParseSpawnBaseline()
	if entity.Number != 35 {
		t.Error("parse baseline, wrong number, got", entity.Number, "want 4")
	}

	got := entity.Marshal()
	got.Reset()
	msg.Reset()

	if diff := cmp.Diff(&msg, got); diff != "" {
		t.Error("baseline marshal error, got diff\n", diff)
	}
}
*/

/*
func TestParseFrame(t *testing.T) {
	msg := Buffer{
		Data: []byte{0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x01, 0x02},
	}
	fr := msg.ParseFrame()
	if fr.Delta != -1 {
		t.Error("parse frame, wrong delta, got", fr.Delta, "want -1")
	}

	got := fr.Marshal()
	got.Reset()
	msg.Reset()

	if diff := cmp.Diff(&msg, got); diff != "" {
		t.Error("frame marshal error, got diff\n", diff)
	}
}
*/

/*
func TestParsePlayerstate(t *testing.T) {
	msg := Buffer{
		Data: []byte{0x00, 0x20, 0x0E, 0x0C, 0x00, 0xF5, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	ps := msg.ParseDeltaPlayerstate(PackedPlayer{})

	got := ps.Marshal()
	got.Reset()
	msg.Reset()

	if diff := cmp.Diff(&msg, got); diff != "" {
		t.Error("frame marshal error, got diff\n", diff)
	}
}
*/

/*
func TestParseDeltaEntities(t *testing.T) {
	msg := Buffer{
		//Buffer: []byte{0x10, 0x01, 0x17, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00},
		Data: []byte{
			0x97, 0x8E, 0x90, 0x0B, 0x01, 0xFF, 0xFF, 0x0D,
			0x00, 0x01, 0xA9, 0x30, 0x8F, 0x1F, 0xC1, 0x20,
			0x04, 0xB6, 0xA9, 0x30, 0x8F, 0x1F, 0xC1, 0x20,
			0x10, 0x18, 0x00, 0x18, 0x00, 0x1A, 0x00, 0x1B,
			0x00, 0x25, 0x00, 0x26, 0x00, 0x29, 0x00, 0x2B,
			0x00, 0x2C, 0x00, 0x2D, 0x00, 0x35, 0x00, 0x36,
			0x00, 0x37, 0x00, 0x38, 0x00, 0x3C, 0x00, 0x3E,
			0x00, 0x3F, 0x00, 0x40, 0x00, 0x41, 0x00, 0x42,
			0x00, 0x43, 0x00, 0x44, 0x00, 0x45, 0x00, 0x46,
			0x00, 0x47, 0x00, 0x49, 0x00, 0x4A, 0x00, 0x4B,
			0x00, 0x52, 0x00, 0x5B, 0x00, 0x5C, 0x00, 0x5D,
			0x00, 0x5E, 0x00, 0x5F, 0x00, 0x77, 0x00, 0x79,
			0x00, 0x00,
		},
	}
	ents := msg.ParsePacketEntities(&ServerFrame{})

	if len(ents) != 36 {
		t.Error("wrong entity count, want 36, got", len(ents))
	}

	got := Buffer{}
	for _, e := range ents {
		got.Append(*e.Marshal())
	}
	got.WriteShort(0)
	got.Reset()
	msg.Reset()

	if diff := cmp.Diff(got, msg); diff != "" {
		t.Error("differences found:\n", diff)
	}
}
*/

/*
func TestRenderSVG(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  []uint8
	}{
		{
			desc:  "TEST 1",
			input: "q2dm1",
			want:  []uint8{0x71, 0x32, 0x64, 0x6d, 0x31, 0x00}, // always ends in null
		},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			msg := MessageBuffer{}
			msg.WriteString(tc.input)
			got := msg.Buffer
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Error("got diff:", diff)
			}
		})
	}
}
*/

/*
func TestValidateSequence(t *testing.T) {
	tests := []struct {
		desc         string
		seq          uint32
		wantreliable bool
		wantsequence uint32
	}{
		{
			desc:         "Test_1 pass",
			seq:          5,
			wantreliable: false,
			wantsequence: 5,
		},
		{
			desc:         "Test_2 pass",
			seq:          2147483652,
			wantreliable: true,
			wantsequence: 4,
		},
	}

	for _, test := range tests {
		r, s := ValidateSequence(test.seq)
		if r != test.wantreliable {
			t.Errorf("%s ValidateSequnce() reliability mismatch - have %t, want %t\n", test.desc, r, test.wantreliable)
		}

		if s != test.wantsequence {
			t.Errorf("%s ValidateSequnce() sequence mismatch - have %d, want %d\n", test.desc, s, test.wantsequence)
		}
	}
}
*/

func TestParseDeltaPlayerstateProto(t *testing.T) {
	tests := []struct {
		name string
		msg  *Buffer
		from *pb.PackedPlayer
		want *pb.PackedPlayer
	}{
		{
			name: "from nil playerstate",
			msg: &Buffer{
				Data: []byte{
					134, 35, 196, 8, 255, 36, 193, 1, 147, 253, 254, 255,
					0, 0, 0, 0, 88, 42, 10, 119, 206, 0, 0, 0, 0, 1, 44,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				},
			},
			from: nil,
			want: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX:   2244,
					OriginY:   9471,
					OriginZ:   449,
					VelocityX: 64915,
					VelocityY: 65534,
				},
				ViewAnglesX: 2602,
				ViewAnglesY: 52855,
				ViewOffsetZ: 88,
				KickAnglesZ: 1,
				GunFrame:    44,
			},
		},
		{
			name: "fov only from valid state",
			msg: &Buffer{
				Data: []byte{
					0, 8, 105, 0, 0, 0, 0,
				},
			},
			from: &pb.PackedPlayer{
				Movestate:   &pb.PlayerMove{},
				ViewAnglesX: 5,
				ViewAnglesY: 5,
				ViewAnglesZ: 5,
				Fov:         90,
			},
			want: &pb.PackedPlayer{
				Movestate:   &pb.PlayerMove{},
				ViewAnglesX: 5,
				ViewAnglesY: 5,
				ViewAnglesZ: 5,
				Fov:         105,
			},
		},
		{
			name: "fov only from valid with movestate",
			msg: &Buffer{
				Data: []byte{
					0, 8, 105, 0, 0, 0, 0,
				},
			},
			from: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 1,
					OriginY: 2,
					OriginZ: 3,
				},
				ViewAnglesX: 5,
				ViewAnglesY: 5,
				ViewAnglesZ: 5,
				Fov:         90,
			},
			want: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 1,
					OriginY: 2,
					OriginZ: 3,
				},
				ViewAnglesX: 5,
				ViewAnglesY: 5,
				ViewAnglesZ: 5,
				Fov:         105,
			},
		},
		{
			name: "fov only from valid with existing stats",
			msg: &Buffer{
				Data: []byte{
					0, 8, 105, 0, 0, 0, 0,
				},
			},
			from: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 1,
					OriginY: 2,
					OriginZ: 3,
				},
				ViewAnglesX: 5,
				ViewAnglesY: 5,
				ViewAnglesZ: 5,
				Fov:         90,
				Stats: map[uint32]int32{
					1:  100,
					5:  50,
					10: 25,
				},
			},
			want: &pb.PackedPlayer{
				Movestate: &pb.PlayerMove{
					OriginX: 1,
					OriginY: 2,
					OriginZ: 3,
				},
				ViewAnglesX: 5,
				ViewAnglesY: 5,
				ViewAnglesZ: 5,
				Fov:         105,
				Stats: map[uint32]int32{
					1:  100,
					5:  50,
					10: 25,
				},
			},
		},
		{
			name: "fov only from valid state with new stats",
			msg: &Buffer{
				Data: []byte{
					0, 8, 105, 96, 0, 0, 0, 9, 0, 0, 0,
				},
			},
			from: &pb.PackedPlayer{
				Movestate:   &pb.PlayerMove{},
				ViewAnglesX: 5,
				ViewAnglesY: 5,
				ViewAnglesZ: 5,
				Fov:         90,
			},
			want: &pb.PackedPlayer{
				Movestate:   &pb.PlayerMove{},
				ViewAnglesX: 5,
				ViewAnglesY: 5,
				ViewAnglesZ: 5,
				Fov:         105,
				Stats: map[uint32]int32{
					5: 9,
					6: 0,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.msg.ParseDeltaPlayerstate(tc.from)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("ParseDeltaPlayerstateProto(%v) = %v, want: %v\n", tc.from, got, tc.want)
			}
		})
	}
}

func TestParseEntityProto(t *testing.T) {
	tests := []struct {
		name string
		msg  *Buffer
		from *pb.PackedEntity
		want *pb.PackedEntity
	}{
		{
			name: "from nil",
			msg: &Buffer{
				Data: []byte{
					148, 4, 2, 71, 2, 239, 27, 3, 69, 196, 8, 255, 36, 2, 0, 0,
				},
			},
			from: nil,
			want: &pb.PackedEntity{
				Number: 2,
				AngleX: 2,
				AngleY: 239,
				Frame:  71,
			},
		},
		{
			name: "no change",
			msg: &Buffer{
				Data: []byte{
					148, 4, 2, 71, 2, 239, 27, 3, 69, 196, 8, 255, 36, 2, 0, 0,
				},
			},
			from: &pb.PackedEntity{
				Number: 2,
				AngleX: 2,
				AngleY: 239,
				Frame:  71,
			},
			want: &pb.PackedEntity{
				Number: 2,
				AngleX: 2,
				AngleY: 239,
				Frame:  71,
			},
		},
		{
			name: "single change",
			msg: &Buffer{
				Data: []byte{
					148, 4, 2, 71, 2, 239, 27, 3, 69, 196, 8, 255, 36, 2, 0, 0,
				},
			},
			from: &pb.PackedEntity{
				Number: 2,
				AngleX: 1,
				AngleY: 239,
				Frame:  71,
			},
			want: &pb.PackedEntity{
				Number: 2,
				AngleX: 2,
				AngleY: 239,
				Frame:  71,
			},
		},
		{
			name: "all new stuff",
			msg: &Buffer{
				Data: []byte{
					148, 4, 2, 71, 2, 239, 27, 3, 69, 196, 8, 255, 36, 2, 0, 0,
				},
			},
			from: &pb.PackedEntity{
				Number:      2,
				AngleZ:      33,
				ModelIndex:  255,
				ModelIndex2: 255,
			},
			want: &pb.PackedEntity{
				Number:      2,
				AngleX:      2,
				AngleY:      239,
				AngleZ:      33,
				ModelIndex:  255,
				ModelIndex2: 255,
				Frame:       71,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bits := tc.msg.ParseEntityBitmask()
			num := tc.msg.ParseEntityNumber(bits)
			got := tc.msg.ParseEntity(tc.from, num, bits)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("ParseEntityProto(%v) = %v, want: %v\n", tc.from, got, tc.want)
			}
		})
	}
}

func TestPlayerstateDiff(t *testing.T) {
	tests := []struct {
		name  string
		data1 string
		data2 string
	}{
		{
			name:  "original",
			data1: "96265F225013C10E5D07F10200000400003654000016FB0AF501FDFFFF000033028000005E000100",
			data2: "96265F225013C10E5D07F10200000400003654000016FB0AF501FDFFFF000033028000005E000100",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b1, err := hex.DecodeString(tc.data1)
			if err != nil {
				t.Fatal(err)
			}
			b2, err := hex.DecodeString(tc.data2)
			if err != nil {
				t.Fatal(err)
			}
			msg1 := NewBuffer(b1)
			msg2 := NewBuffer(b2)
			ps1 := msg1.ParseDeltaPlayerstate(nil)
			ps2 := msg2.ParseDeltaPlayerstate(nil)
			if diff := cmp.Diff(ps1, ps2, protocmp.Transform()); diff != "" {
				t.Errorf("Playerstate diff: (+got/-want):\n%s", diff)
			}
		})
	}
}

func TestPacketEntitiesDump(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "original",
			data: "B3828001013A5F225013C10E5F225013C10E04000003015F225013C10E0054000000",
		},
		{
			name: "new",
			data: "B302013A5F225013C10E04000003015F225013C10E00",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			msg := NewBuffer(bytes)
			ents := msg.ParsePacketEntities(nil)
			for _, ent := range ents {
				t.Error("\n" + prototext.Format(ent))
			}
		})
	}
}

func TestParsePacketDump(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "original",
			data: "14100000000F0000000001021196265F225013C10E5D07F10200000400003654000016FB0AF501FDFFFF000033028000005E00010012B3828001013A5F225013C10E5F225013C10E04000003015F225013C10E00",
		},
		{
			name: "test1",
			data: "0A03636C616972653A203A29200A00141100000010000000000102118626D9227F13C10E3404AC01000000004138000017FB0AF502FCFEFF0000230080000000001293808001013BD9227F13D9227F13C10E0000",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			msg := NewBuffer(bytes)
			packet, err := msg.ParsePacket(nil)
			if err != nil {
				t.Error(err)
			}
			t.Error("\n" + prototext.Format(packet))
		})
	}
}
