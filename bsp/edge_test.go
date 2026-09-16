package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseEdges(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetEdges()), 1795; got != want {
		t.Errorf("edge count - have: %d, want: %d", got, want)
	}
	if got, want := len(bsp.GetSurfEdges()), 3588; got != want {
		t.Errorf("surfedge count - have: %d, want: %d", got, want)
	}
}

func TestEdgeRoundTrip(t *testing.T) {
	want := []*bpb.BSPEdge{
		{Vertex1: 0, Vertex2: 1},
		{Vertex1: 65535, Vertex2: 2},
	}
	data := marshalEdges(want)
	if len(data) != len(want)*edgeSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*edgeSize)
	}
	got, err := parseEdges(data)
	if err != nil {
		t.Fatal(err)
	}
	for i := range want {
		if got[i].GetVertex1() != want[i].GetVertex1() || got[i].GetVertex2() != want[i].GetVertex2() {
			t.Errorf("edge %d - have: %+v, want: %+v", i, got[i], want[i])
		}
	}
}

func TestSurfEdgeRoundTrip(t *testing.T) {
	want := []int32{5, -5, 0, -1}
	data := marshalSurfEdges(want)
	got, err := parseSurfEdges(data)
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
