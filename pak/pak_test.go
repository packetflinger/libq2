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

func TestAddFile(t *testing.T) {
	pak, e := OpenPakFile("../testdata/test.pak")
	if e != nil {
		t.Errorf("%v", e)
	}

	e = pak.AddFile("../testdata/testfile.txt")
	if e != nil {
		t.Error(e)
	}

	if len(pak.Files) != 3 {
		t.Error("Wrong file count")
	}

	if string(pak.Files[len(pak.Files)-1].Data) != "test\n" {
		t.Error("new file contents mismatch")
	}
	pak.Close()
}

func TestRemoveFile(t *testing.T) {
	pak, e := OpenPakFile("../testdata/test.pak")
	if e != nil {
		t.Errorf("%v", e)
	}

	e = pak.RemoveFile("test2.cfg")
	if e != nil {
		t.Error(e)
	}

	if len(pak.Files) != 1 {
		t.Error("Wrong file count")
	}

	pak.Close()
}
