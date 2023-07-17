package pak

import (
	"os"
	"testing"
)

func TestListFiles(t *testing.T) {
	infile := "../testdata/test.pkz"
	pkz, e := OpenPKZFile(infile)
	if e != nil {
		t.Errorf("%v", e)
	}

	files, err := pkz.ListFiles()
	if err != nil {
		t.Error(err)
	}
	if len(files) != 3 {
		t.Error("Incorrect file count, have:", len(files), ", want: 3")
	}
	pkz.Close()
}

func TestPKZAddFile(t *testing.T) {
	// make a copy so we don't ruin our test file
	originalInput := "../testdata/test.pkz"
	tempfile := "../testdata/temp-test.pkz"
	data, err := os.ReadFile(originalInput)
	if err != nil {
		t.Errorf("%v", err)
	}
	newInput, err := os.Create(tempfile)
	if err != nil {
		t.Errorf("%v", err)
	}
	newInput.Write(data)
	newInput.Close()

	pkz, e := OpenPKZFile(tempfile)
	if e != nil {
		pkz.Close()
		os.Remove(tempfile)
		t.Errorf("%v", e)
	}

	filesbefore, err := pkz.ListFiles()
	if err != nil {
		pkz.Close()
		os.Remove(tempfile)
		t.Error(err)
	}

	if len(filesbefore) != 3 {
		pkz.Close()
		os.Remove(tempfile)
		t.Error("Incorrect file count, have:", len(filesbefore), ", want: 3")
	}

	err = pkz.AddFile("../testdata/test.dm2")
	if err != nil {
		pkz.Close()
		os.Remove(tempfile)
		t.Errorf("%v", err)
	}

	filesafter, err := pkz.ListFiles()
	if err != nil {
		pkz.Close()
		os.Remove(tempfile)
		t.Error(err)
	}

	if len(filesafter) != 4 {
		pkz.Close()
		os.Remove(tempfile)
		t.Error("Incorrect file count, have:", len(filesafter), ", want: 4")
	}
	pkz.Close()
}
