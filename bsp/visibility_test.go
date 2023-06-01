package bsp

import (
	"testing"
)

func TestFetchVisibility(t *testing.T) {
	infile := "../testdata/backup.bsp"
	bsp, e := OpenBSPFile(infile)
	if e != nil {
		t.Errorf("%v", e)
	}

	vis := bsp.FetchVisibility()
	if count := len(vis); count != 768 {
		t.Errorf("Wrong visibility count, got %d, want 768\n", count)
	}
}
