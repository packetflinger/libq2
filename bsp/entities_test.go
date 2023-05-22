package bsp

import (
	"testing"
)

func TestFetchEntityString(t *testing.T) {
	bsp, err := OpenBSPFile("../testdata/backup.bsp")
	if err != nil {
		t.Error(err)
	}

	entstr := bsp.FetchEntityString()
	if len(entstr) == 0 {
		t.Error("zero length entity string")
	}
}

func TestFetchEntities(t *testing.T) {
	bsp, err := OpenBSPFile("../testdata/backup.bsp")
	if err != nil {
		t.Error(err)
	}

	ents := bsp.FetchEntities()
	if len(ents) < 1 {
		t.Error("zero entities fetched")
	}
}

func TestBuildEntityString(t *testing.T) {
	bsp, err := OpenBSPFile("../testdata/backup.bsp")
	if err != nil {
		t.Error(err)
	}

	str := bsp.BuildEntityString()
	if len(str) < 1 {
		t.Error("zero entities fetched")
	}
}
