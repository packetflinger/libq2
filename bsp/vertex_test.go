package bsp

import (
	"testing"
)

func TestFetchVertices(t *testing.T) {
	infile := "../testdata/backup.bsp"
	bsp, e := OpenBSPFile(infile)
	if e != nil {
		t.Errorf("%v", e)
	}

	v := bsp.FetchVertices()
	if len(v) != 1054 {
		t.Errorf("Wrong plane count, want 402, have %d\n", len(v))
	}
}
