package message

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestServerData(t *testing.T) {
	msg := MessageBuffer{
		Buffer: []byte{
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

func TestParseConfigstring(t *testing.T) {
	msg := MessageBuffer{
		Buffer: []byte{
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

func TestParseBaseline(t *testing.T) {
	msg := MessageBuffer{
		Buffer: []byte{0x83, 0xCA, 0x04, 0x23, 0x1C, 0x01, 0x00, 0x02, 0x80, 0x21, 0x00, 0x28, 0x78, 0x1C},
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

func TestParseFrame(t *testing.T) {
	msg := MessageBuffer{
		Buffer: []byte{0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x01, 0x02},
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

func TestParsePlayerstate(t *testing.T) {
	msg := MessageBuffer{
		Buffer: []byte{0x00, 0x20, 0x0E, 0x0C, 0x00, 0xF5, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	ps := msg.ParseDeltaPlayerstate(PackedPlayer{})

	got := ps.Marshal()
	got.Reset()
	msg.Reset()

	if diff := cmp.Diff(&msg, got); diff != "" {
		t.Error("frame marshal error, got diff\n", diff)
	}
}

func TestParseDeltaEntities(t *testing.T) {
	msg := MessageBuffer{
		//Buffer: []byte{0x10, 0x01, 0x17, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00},
		Buffer: []byte{
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

	got := MessageBuffer{}
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
