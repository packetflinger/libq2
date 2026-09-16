package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseVisibility(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	vis := bsp.GetVisibility()
	if got, want := vis.GetClusterCount(), int32(140); got != want {
		t.Errorf("cluster count - have: %d, want: %d", got, want)
	}
	if got, want := len(vis.GetOffsets()), 140; got != want {
		t.Errorf("offset count - have: %d, want: %d", got, want)
	}
}

func TestVisibilityRoundTrip(t *testing.T) {
	want := &bpb.BSPVisibility{
		ClusterCount: 2,
		Offsets: []*bpb.BSPVisOffset{
			{Pvs: 20, Phs: 24},
			{Pvs: 28, Phs: 32},
		},
		BitVectors: []byte{0xff, 0x00, 0xab, 0xcd},
	}
	data := marshalVisibility(want)
	got, err := parseVisibility(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetClusterCount() != want.GetClusterCount() {
		t.Errorf("cluster count - have: %d, want: %d", got.GetClusterCount(), want.GetClusterCount())
	}
	if len(got.GetOffsets()) != len(want.GetOffsets()) {
		t.Fatalf("offset count - have: %d, want: %d", len(got.GetOffsets()), len(want.GetOffsets()))
	}
	for i := range want.GetOffsets() {
		if got.GetOffsets()[i].GetPvs() != want.GetOffsets()[i].GetPvs() ||
			got.GetOffsets()[i].GetPhs() != want.GetOffsets()[i].GetPhs() {
			t.Errorf("offset %d - have: %v, want: %v", i, got.GetOffsets()[i], want.GetOffsets()[i])
		}
	}
	if string(got.GetBitVectors()) != string(want.GetBitVectors()) {
		t.Errorf("bit vectors - have: %v, want: %v", got.GetBitVectors(), want.GetBitVectors())
	}
}

func TestVisibilityEmptyLump(t *testing.T) {
	got, err := parseVisibility(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetClusterCount() != 0 {
		t.Errorf("cluster count - have: %d, want: 0", got.GetClusterCount())
	}
	if data := marshalVisibility(got); len(data) != 0 {
		t.Errorf("marshaled empty visibility - have: %d bytes, want: 0", len(data))
	}
}
