package bsp

import (
	"strings"
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseTexInfo(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetTextureInfo()), 63; got != want {
		t.Errorf("texinfo count - have: %d, want: %d", got, want)
	}
}

func TestTexInfoRoundTrip(t *testing.T) {
	want := []*bpb.BSPTexInfo{
		{
			SAxis:       &bpb.Vector3{X: 1, Y: 0, Z: 0},
			SOffset:     16,
			TAxis:       &bpb.Vector3{X: 0, Y: -1, Z: 0},
			TOffset:     -32.5,
			Flags:       3,
			Value:       0,
			Texture:     "e1u1/wall",
			NextTexinfo: -1,
		},
	}
	data, err := marshalTexInfo(want)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(want)*texInfoSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*texInfoSize)
	}
	got, err := parseTexInfo(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("parsed count - have: %d, want: 1", len(got))
	}
	if got[0].GetTexture() != want[0].GetTexture() {
		t.Errorf("texture - have: %q, want: %q", got[0].GetTexture(), want[0].GetTexture())
	}
	if got[0].GetSOffset() != want[0].GetSOffset() || got[0].GetTOffset() != want[0].GetTOffset() {
		t.Errorf("offsets - have: (%v, %v), want: (%v, %v)", got[0].GetSOffset(), got[0].GetTOffset(), want[0].GetSOffset(), want[0].GetTOffset())
	}
	if got[0].GetNextTexinfo() != want[0].GetNextTexinfo() {
		t.Errorf("next_texinfo - have: %d, want: %d", got[0].GetNextTexinfo(), want[0].GetNextTexinfo())
	}
}

func TestTexInfoNameTooLong(t *testing.T) {
	longName := strings.Repeat("x", texInfoNameSize)
	_, err := marshalTexInfo([]*bpb.BSPTexInfo{{Texture: longName}})
	if err == nil {
		t.Error("expected an error for a texture name that doesn't fit in the fixed field")
	}
}
