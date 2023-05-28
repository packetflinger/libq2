package demo

import (
	"fmt"
	"testing"

	m "github.com/packetflinger/libq2/message"
)

func TestOpenDM2File(t *testing.T) {
	demo, e := OpenDM2File("../testdata/test.dm2")
	if e != nil {
		t.Errorf("%v", e)
	}

	if demo.Handle == nil {
		t.Error("Handle - file handle is nil")
	}

	demo.Close()
}

func TestParseDM2(t *testing.T) {
	pmsg := ""
	demo, _ := OpenDM2File("../testdata/test.dm2")
	cb := m.MessageCallbacks{
		Print: func(p *m.Print) {
			pmsg = p.String
			fmt.Println(p.Level, p.String)
		},
	}

	err := demo.ParseDM2(cb)
	if err != nil {
		t.Error(err)
	}

	if pmsg != "claire: hi\n" {
		t.Errorf("Print msg not expected")
	}
	demo.Close()
}
