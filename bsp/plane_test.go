package bsp

import (
	"testing"
)

func TestParsePlanes(t *testing.T) {
	infile := "../testdata/backup.bsp"
	bsp, e := OpenBSPFile(infile)
	if e != nil {
		t.Errorf("%v", e)
	}

	planes := bsp.FetchPlanes()
	if len(planes) != 402 {
		t.Errorf("Wrong plane count, want 402, have %d\n", len(planes))
	}
}
