package demo

import (
	"fmt"
	"testing"

	m "github.com/packetflinger/libq2/message"
)

func TestNewDM2Demo(t *testing.T) {
	tests := []struct {
		name string
		file string
		want int
	}{
		{
			name: "Test 1",
			file: "../testdata/test.dm2",
			want: 7931,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			demo, err := NewDM2Demo(tc.file)
			if err != nil {
				t.Error(err)
			}
			got := len(demo.binaryData)
			if got != tc.want {
				t.Errorf("\nSize mismatch\ngot: %d, want: %d\n", got, tc.want)
			}
		})
	}
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

func TestWrite(t *testing.T) {
	demo, _ := OpenDM2File("../testdata/claire-shloo_PFNJ_q2dm1_20230505-154051.dm2")
	cb := m.MessageCallbacks{
		Frame: func(fr *m.FrameMsg) {
			fmt.Println(fr.Number, fr.Delta)
		},
	}

	err := demo.ParseDM2(cb)
	if err != nil {
		t.Error(err)
	}

	psfrom := demo.Frames[1000].Playerstate
	psto := demo.Frames[2000].Playerstate
	msg := m.MessageBuffer{}
	msg.WriteDeltaPlayerstate(&psto, &psfrom)

	//fmt.Println(psfrom)
	//fmt.Println("ljsfljsf")
	//fmt.Println(psto)
	//fmt.Println("lajksflk")
	//fmt.Println(msg)
	//demo.Write()
	//t.Error()
	demo.Close()
}

func TestSliceCopyOK(t *testing.T) {
	original := []string{
		"alice",
		"bob",
		"charlie",
	}
	copy := append([]string{}, original...)
	copy[1] = "joe"

	fmt.Println(original)
	fmt.Println(copy)
	t.Error()
}
