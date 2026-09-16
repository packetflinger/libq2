package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseLeaves(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetLeaves()), 360; got != want {
		t.Errorf("leaf count - have: %d, want: %d", got, want)
	}
	if got, want := len(bsp.GetLeafFaces()), 870; got != want {
		t.Errorf("leaf face table count - have: %d, want: %d", got, want)
	}
	if got, want := len(bsp.GetLeafBrushes()), 340; got != want {
		t.Errorf("leaf brush table count - have: %d, want: %d", got, want)
	}
}

func TestLeafRoundTrip(t *testing.T) {
	want := []*bpb.BSPLeaf{
		{
			Contents: 1,
			Cluster:  5,
			Area:     2,
			Bounds: &bpb.BoundingBox{
				Mins: &bpb.Vector3{X: -10, Y: -20, Z: -30},
				Maxs: &bpb.Vector3{X: 10, Y: 20, Z: 30},
			},
			FirstLeafFace:  1,
			NumLeafFaces:   2,
			FirstLeafBrush: 3,
			NumLeafBrushes: 4,
		},
	}
	data := marshalLeaves(want)
	if len(data) != len(want)*leafSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*leafSize)
	}
	got, err := parseLeaves(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("parsed count - have: %d, want: 1", len(got))
	}
	l, w := got[0], want[0]
	if l.GetContents() != w.GetContents() || l.GetCluster() != w.GetCluster() || l.GetArea() != w.GetArea() {
		t.Errorf("leaf fields - have: %+v, want: %+v", l, w)
	}
	if l.GetFirstLeafFace() != w.GetFirstLeafFace() || l.GetNumLeafFaces() != w.GetNumLeafFaces() ||
		l.GetFirstLeafBrush() != w.GetFirstLeafBrush() || l.GetNumLeafBrushes() != w.GetNumLeafBrushes() {
		t.Errorf("leaf ranges - have: %+v, want: %+v", l, w)
	}
}

func TestUint16TableRoundTrip(t *testing.T) {
	want := []uint32{0, 1, 65535, 42}
	data := marshalUint16Table(want)
	got, err := parseUint16Table(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("count - have: %d, want: %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d - have: %d, want: %d", i, got[i], want[i])
		}
	}
}
