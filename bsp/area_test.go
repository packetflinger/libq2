package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseAreas(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(bsp.GetAreas()), 2; got != want {
		t.Errorf("area count - have: %d, want: %d", got, want)
	}
	if got, want := len(bsp.GetAreaPortals()), 1; got != want {
		t.Errorf("area portal count - have: %d, want: %d", got, want)
	}
}

func TestAreaRoundTrip(t *testing.T) {
	want := []*bpb.BSPArea{
		{NumAreaPortals: 1, FirstAreaPortal: 0},
	}
	data := marshalAreas(want)
	if len(data) != len(want)*areaSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*areaSize)
	}
	got, err := parseAreas(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].GetNumAreaPortals() != want[0].GetNumAreaPortals() ||
		got[0].GetFirstAreaPortal() != want[0].GetFirstAreaPortal() {
		t.Errorf("area round trip - have: %+v, want: %+v", got, want)
	}
}

func TestAreaPortalRoundTrip(t *testing.T) {
	want := []*bpb.BSPAreaPortal{
		{PortalNum: 1, OtherArea: 2},
	}
	data := marshalAreaPortals(want)
	if len(data) != len(want)*areaPortalSize {
		t.Fatalf("marshaled size - have: %d, want: %d", len(data), len(want)*areaPortalSize)
	}
	got, err := parseAreaPortals(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].GetPortalNum() != want[0].GetPortalNum() || got[0].GetOtherArea() != want[0].GetOtherArea() {
		t.Errorf("area portal round trip - have: %+v, want: %+v", got, want)
	}
}
