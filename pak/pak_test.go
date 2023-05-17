package pak

import (
	"testing"
)

func TestOpenPakFile(t *testing.T) {
	infile := "../testdata/test.pak"
	pak, e := OpenPakFile(infile)
	if e != nil {
		t.Errorf("%v", e)
	}

	if pak.Handle == nil {
		t.Error("Handle - file handle is nil")
	}
	if len(pak.Files) != 2 {
		t.Error("Incorrect file count, have:", len(pak.Files), ", want: 2")
	}
	pak.Close()
}
