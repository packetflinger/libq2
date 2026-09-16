package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseVertices(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetVertices()), 1054; got != want {
		t.Errorf("vertex count - have: %d, want: %d", got, want)
	}
}

func TestVertexRoundTrip(t *testing.T) {
	want := []*bpb.BSPVertex{
		{Point: &bpb.Vector3{X: 1, Y: 2, Z: 3}},
		{Point: &bpb.Vector3{X: -100.5, Y: 0, Z: 200.25}},
	}
	data := marshalVertices(want)
	if len(data) != len(want)*vertexSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*vertexSize)
	}
	got, err := parseVertices(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("parsed count - have: %d, want: %d", len(got), len(want))
	}
	for i := range want {
		if got[i].GetPoint().GetX() != want[i].GetPoint().GetX() ||
			got[i].GetPoint().GetY() != want[i].GetPoint().GetY() ||
			got[i].GetPoint().GetZ() != want[i].GetPoint().GetZ() {
			t.Errorf("vertex %d - have: %v, want: %v", i, got[i].GetPoint(), want[i].GetPoint())
		}
	}
}

func TestParseVerticesBadLength(t *testing.T) {
	if _, err := parseVertices(make([]byte, vertexSize+1)); err == nil {
		t.Error("expected an error for a lump size that isn't a multiple of the record size")
	}
}
