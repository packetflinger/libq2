package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseFaces(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetFaces()), 722; got != want {
		t.Errorf("face count - have: %d, want: %d", got, want)
	}
	for i, f := range bsp.GetFaces() {
		if len(f.GetLightStyles()) != faceLightStyles {
			t.Fatalf("face %d has %d light styles, want %d", i, len(f.GetLightStyles()), faceLightStyles)
		}
	}
}

func TestFaceRoundTrip(t *testing.T) {
	want := []*bpb.BSPFace{
		{
			PlaneNum:    12,
			Side:        1,
			FirstEdge:   100,
			NumEdges:    4,
			Texinfo:     3,
			LightStyles: []uint32{0, 255, 255, 255},
			LightOffset: 4096,
		},
	}
	data, err := marshalFaces(want)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(want)*faceSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*faceSize)
	}
	got, err := parseFaces(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("parsed count - have: %d, want: 1", len(got))
	}
	f, w := got[0], want[0]
	if f.GetPlaneNum() != w.GetPlaneNum() || f.GetSide() != w.GetSide() || f.GetFirstEdge() != w.GetFirstEdge() ||
		f.GetNumEdges() != w.GetNumEdges() || f.GetTexinfo() != w.GetTexinfo() || f.GetLightOffset() != w.GetLightOffset() {
		t.Errorf("face fields - have: %+v, want: %+v", f, w)
	}
	for i := range w.GetLightStyles() {
		if f.GetLightStyles()[i] != w.GetLightStyles()[i] {
			t.Errorf("light style %d - have: %d, want: %d", i, f.GetLightStyles()[i], w.GetLightStyles()[i])
		}
	}
}

func TestFaceWrongLightStyleCount(t *testing.T) {
	_, err := marshalFaces([]*bpb.BSPFace{{LightStyles: []uint32{0, 0}}})
	if err == nil {
		t.Error("expected an error for a face with the wrong number of light styles")
	}
}
