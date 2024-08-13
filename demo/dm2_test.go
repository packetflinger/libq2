package demo

import (
	"testing"
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
		{
			name:           "Test 2",
			fileIn:         "../testdata/testduel.dm2",
			wantFrameCount: 3199,
		},
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
			got := len(demo.textProto.GetFrames())
			if got != tc.wantFrameCount {
				t.Errorf("wrong frame count, got %d, want %d\n", got, tc.wantFrameCount)
			}
		})
	}
}

func TestNextLump(t *testing.T) {
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
			demo := &DM2Demo{
				binaryData: tc.data,
			}
			_, got, err := demo.nextLump()
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

/*
func TestWrite(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{
			name: "Test 1",
			file: "../testdata/test.dm2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			demo, err := NewDM2Demo(tc.file)
			if err != nil {
				t.Error(err)
			}
			t.Error(demo.binaryData[7711:7747])
		})
	}
}
*/
