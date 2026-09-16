package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseBrushes(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetBrushes()), 248; got != want {
		t.Errorf("brush count - have: %d, want: %d", got, want)
	}
	if got, want := len(bsp.GetBrushSides()), 1528; got != want {
		t.Errorf("brush side count - have: %d, want: %d", got, want)
	}
}

func TestBrushRoundTrip(t *testing.T) {
	want := []*bpb.BSPBrush{
		{FirstSide: 4, NumSides: 6, Contents: 1},
	}
	data := marshalBrushes(want)
	if len(data) != len(want)*brushSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*brushSize)
	}
	got, err := parseBrushes(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].GetFirstSide() != want[0].GetFirstSide() ||
		got[0].GetNumSides() != want[0].GetNumSides() || got[0].GetContents() != want[0].GetContents() {
		t.Errorf("brush round trip - have: %+v, want: %+v", got, want)
	}
}

func TestBrushSideRoundTrip(t *testing.T) {
	want := []*bpb.BSPBrushSide{
		{PlaneNum: 3, Texinfo: -1},
	}
	data := marshalBrushSides(want)
	if len(data) != len(want)*brushSideSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*brushSideSize)
	}
	got, err := parseBrushSides(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].GetPlaneNum() != want[0].GetPlaneNum() || got[0].GetTexinfo() != want[0].GetTexinfo() {
		t.Errorf("brush side round trip - have: %+v, want: %+v", got, want)
	}
}
