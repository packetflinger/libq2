package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParsePlanes(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetPlanes()), 402; got != want {
		t.Errorf("plane count - have: %d, want: %d", got, want)
	}
}

func TestPlaneRoundTrip(t *testing.T) {
	want := []*bpb.BSPPlane{
		{Normal: &bpb.Vector3{X: 1, Y: 0, Z: 0}, Distance: 64, Type: 0},
		{Normal: &bpb.Vector3{X: 0, Y: 0.5, Z: -1}, Distance: -128.5, Type: 3},
	}
	data, err := marshalPlanes(want)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(want)*planeSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*planeSize)
	}
	got, err := parsePlanes(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("parsed count - have: %d, want: %d", len(got), len(want))
	}
	for i := range want {
		if got[i].GetNormal().GetX() != want[i].GetNormal().GetX() ||
			got[i].GetNormal().GetY() != want[i].GetNormal().GetY() ||
			got[i].GetNormal().GetZ() != want[i].GetNormal().GetZ() {
			t.Errorf("plane %d normal - have: %v, want: %v", i, got[i].GetNormal(), want[i].GetNormal())
		}
		if got[i].GetDistance() != want[i].GetDistance() {
			t.Errorf("plane %d distance - have: %v, want: %v", i, got[i].GetDistance(), want[i].GetDistance())
		}
		if got[i].GetType() != want[i].GetType() {
			t.Errorf("plane %d type - have: %v, want: %v", i, got[i].GetType(), want[i].GetType())
		}
	}
}

func TestParsePlanesBadLength(t *testing.T) {
	if _, err := parsePlanes(make([]byte, planeSize+1)); err == nil {
		t.Error("expected an error for a lump size that isn't a multiple of the record size")
	}
}
