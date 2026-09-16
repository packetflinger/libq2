package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseNodes(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetNodes()), 357; got != want {
		t.Errorf("node count - have: %d, want: %d", got, want)
	}
}

func TestNodeRoundTrip(t *testing.T) {
	want := []*bpb.BSPNode{
		{
			PlaneNum:   5,
			FrontChild: 6,
			BackChild:  -1,
			Bounds: &bpb.BoundingBox{
				Mins: &bpb.Vector3{X: -100, Y: -200, Z: -50},
				Maxs: &bpb.Vector3{X: 100, Y: 200, Z: 50},
			},
			FirstFace: 10,
			NumFaces:  4,
		},
	}
	data := marshalNodes(want)
	if len(data) != len(want)*nodeSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*nodeSize)
	}
	got, err := parseNodes(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("parsed count - have: %d, want: 1", len(got))
	}
	n := got[0]
	w := want[0]
	if n.GetPlaneNum() != w.GetPlaneNum() || n.GetFrontChild() != w.GetFrontChild() || n.GetBackChild() != w.GetBackChild() {
		t.Errorf("node fields - have: %+v, want: %+v", n, w)
	}
	if n.GetBounds().GetMins().GetX() != w.GetBounds().GetMins().GetX() ||
		n.GetBounds().GetMaxs().GetZ() != w.GetBounds().GetMaxs().GetZ() {
		t.Errorf("node bounds - have: %v, want: %v", n.GetBounds(), w.GetBounds())
	}
	if n.GetFirstFace() != w.GetFirstFace() || n.GetNumFaces() != w.GetNumFaces() {
		t.Errorf("node face range - have: (%d,%d), want: (%d,%d)", n.GetFirstFace(), n.GetNumFaces(), w.GetFirstFace(), w.GetNumFaces())
	}
}
