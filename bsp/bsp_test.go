package bsp

import (
	"testing"
)

func TestOpenBSPFile(t *testing.T) {
	infile := "../testdata/backup.bsp"
	bsp, e := OpenBSPFile(infile)
	if e != nil {
		t.Errorf("%v", e)
	}

	if bsp.Name != "backup" {
		t.Error("Name - have:", bsp.Name, "want: backup")
	}
	if bsp.Handle == nil {
		t.Error("Handle - file handle is nil")
	}
	if bsp.Header.Length != HeaderLen {
		t.Error("Header Length - have:", bsp.Header.Length, "want:", HeaderLen)
	}

	if !bsp.Validate() {
		t.Error("Not valid BSP")
	}
	bsp.Close()
}
