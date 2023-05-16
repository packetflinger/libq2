package bsp

import (
	"testing"
)

func TestFetchTextures(t *testing.T) {
	infile := "../testdata/backup.bsp"
	bsp, e := OpenBSPFile(infile)
	if e != nil {
		t.Errorf("%v", e)
	}

	textures := bsp.FetchTextures()
	if len(textures) != 63 {
		t.Error("Wrong texture count")
	}
}
