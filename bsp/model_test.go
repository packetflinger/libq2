package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseModels(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetModels()), 2; got != want {
		t.Errorf("model count - have: %d, want: %d", got, want)
	}
}

func TestModelRoundTrip(t *testing.T) {
	want := []*bpb.BSPModel{
		{
			Bounds: &bpb.BoundingBox{
				Mins: &bpb.Vector3{X: -512.5, Y: -256, Z: -64},
				Maxs: &bpb.Vector3{X: 512.5, Y: 256, Z: 64},
			},
			Origin:    &bpb.Vector3{X: 1, Y: 2, Z: 3},
			HeadNode:  7,
			FirstFace: 20,
			NumFaces:  30,
		},
	}
	data := marshalModels(want)
	if len(data) != len(want)*modelSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*modelSize)
	}
	got, err := parseModels(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("parsed count - have: %d, want: 1", len(got))
	}
	m, w := got[0], want[0]
	if m.GetHeadNode() != w.GetHeadNode() || m.GetFirstFace() != w.GetFirstFace() || m.GetNumFaces() != w.GetNumFaces() {
		t.Errorf("model fields - have: %+v, want: %+v", m, w)
	}
	if m.GetBounds().GetMins().GetX() != w.GetBounds().GetMins().GetX() ||
		m.GetBounds().GetMaxs().GetY() != w.GetBounds().GetMaxs().GetY() {
		t.Errorf("model bounds - have: %v, want: %v", m.GetBounds(), w.GetBounds())
	}
	if m.GetOrigin().GetZ() != w.GetOrigin().GetZ() {
		t.Errorf("model origin - have: %v, want: %v", m.GetOrigin(), w.GetOrigin())
	}
}
