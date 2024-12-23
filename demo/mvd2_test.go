package demo

import (
	"testing"
)

func TestOpenMVD2File(t *testing.T) {
	demo, e := OpenMVD2File("../testdata/test.mvd2")
	if e != nil {
		t.Errorf("%v", e)
	}

	if demo.Handle == nil {
		t.Error("Handle - file handle is nil")
	}

	demo.Close()
}

/*
func TestParseMVD2(t *testing.T) {
	demo, _ := OpenMVD2File("../testdata/test.mvd2")
	cb := message.Callback{}

	err := demo.Parse(cb)
	if err != nil {
		t.Error(err)
	}

	demo.Close()
}
*/
