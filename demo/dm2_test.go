package demo

import (
	"os"
	"testing"

	"google.golang.org/protobuf/encoding/prototext"
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

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name           string
		fileIn         string
		wantFrameCount int
	}{
		{
			name:           "Test 1",
			fileIn:         "../testdata/test.dm2",
			wantFrameCount: 23,
		},
		/*{
			name:           "Test 2",
			fileIn:         "../testdata/testduel.dm2",
			wantFrameCount: 3199,
		},*/
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			demo, err := NewDM2Demo(tc.fileIn)
			if err != nil {
				t.Error(err)
			}
			err = demo.Unmarshal()
			if err != nil {
				t.Error(err)
			}
			t.Error(prototext.Format(demo.textProto))
		})
	}
}

func TestNextPacket(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{
			name: "Test 1",
			data: []byte{
				32, 0, 0, 0,
				20,
				18, 0, 0, 0,
				17, 0, 0, 0,
				0,
				1,
				2,
				17,
				0, 32,
				20, 12, 0, 245, 0, 0, 0, 0, 0, 0, 0,
				18,
				16,
				1, 30, 0, 0},
			want: 32,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			demo := &DM2Parser{
				binaryData: tc.data,
			}
			_, got, err := demo.NextPacket()
			if err != nil {
				t.Error(err)
			}
			if got != tc.want {
				t.Errorf("\nSize mismatch\ngot: %d, want: %d\n", got, tc.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		inFile  string
		outFile string
	}{
		{
			name:    "Test 1",
			inFile:  "../testdata/test.dm2",
			outFile: "../testdata/output-test.dm2.pb",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			demo, err := NewDM2Demo(tc.inFile)
			if err != nil {
				t.Error(err)
			}
			err = demo.Unmarshal()
			if err != nil {
				t.Error(err)
			}
			//demo.WriteTextProto(tc.outFile)
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		inFile  string
		outFile string
	}{
		/*{
			name:    "multiple entities spinning",
			inFile:  "/Users/joe/.quake2/baseq2/demos/test-dm1ents.dm2",
			outFile: "/Users/joe/.quake2/baseq2/demos/test-dm1ents-out.dm2",
		},
		{
			name:    "blaster shot with sound",
			inFile:  "/Users/joe/.quake2/baseq2/demos/test-dm1blaster.dm2",
			outFile: "/Users/joe/.quake2/baseq2/demos/test-dm1blaster-out.dm2",
		},
		{
			name:    "picking up ssh and shells one shell doesn't disappear",
			inFile:  "/Users/joe/.quake2/baseq2/demos/test-dm1pickup.dm2",
			outFile: "/Users/joe/.quake2/baseq2/demos/test-dm1pickup-out.dm2",
		},*/
		{
			name:    "fall and land with sound health highlight",
			inFile:  "/Users/joe/.quake2/baseq2/demos/test-dm1fall2.dm2",
			outFile: "/Users/joe/.quake2/baseq2/demos/test-dm1fall2-out.dm2",
		},
		/*{
			name:    "rocket shot with explosion",
			inFile:  "/Users/joe/.quake2/baseq2/demos/test-dm1rocket.dm2",
			outFile: "/Users/joe/.quake2/baseq2/demos/test-dm1rocket-out.dm2",
		},*/
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			demo, err := NewDM2Demo(tc.inFile)
			if err != nil {
				t.Error(err)
			}
			err = demo.Unmarshal()
			if err != nil {
				t.Error(err)
			}
			data, err := demo.Marshal()
			if err != nil {
				t.Error(err)
			}
			err = os.WriteFile(tc.outFile, data, 0777)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDemoDebug(t *testing.T) {
	tests := []struct {
		name   string
		inFile string
	}{

		/*{
			name:   "fall and land with sound health highlight",
			inFile: "/Users/joe/.quake2/baseq2/demos/test-dm1fall2.dm2",
		},*/

		{
			name:   "rocket shot with explosion",
			inFile: "/Users/joe/.quake2/baseq2/demos/test-dm1fall2-out.dm2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			demo, err := NewDM2Demo(tc.inFile)
			if err != nil {
				t.Error(err)
			}
			err = demo.Unmarshal()
			if err != nil {
				t.Error(err)
			}
			t.Error(prototext.Format(demo.textProto))
		})
	}
}
