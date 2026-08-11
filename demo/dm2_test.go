package demo

import (
	"os"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name              string
		fileIn            string
		wantBaselines     int
		wantConfigstrings int
		wantFrames        int
	}{
		{
			name:              "Test 1",
			fileIn:            "../testdata/test.dm2",
			wantBaselines:     107,
			wantConfigstrings: 231,
			wantFrames:        23,
		},
		{
			name:              "Test 2",
			fileIn:            "../testdata/testduel.dm2",
			wantBaselines:     70,
			wantConfigstrings: 241,
			wantFrames:        3199,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(tc.fileIn)
			if err != nil {
				t.Fatal("opening demo file:", content)
			}
			demo := NewDM2Parser()
			err = demo.Unmarshal(content)
			if err != nil {
				t.Error(err)
			}
			if len(demo.textProto.GetBaselines()) != tc.wantBaselines {
				t.Errorf("Unmarshal() baseline mismatch: got %d, want %d\n", len(demo.textProto.Baselines), tc.wantBaselines)
			}
			if len(demo.textProto.GetConfigstrings()) != tc.wantConfigstrings {
				t.Errorf("Unmarshal() configstring mismatch: got %d, want %d\n", len(demo.textProto.GetConfigstrings()), tc.wantConfigstrings)
			}
			if len(demo.textProto.GetFrames()) != tc.wantFrames {
				t.Errorf("Unmarshal() frame mismatch: got %d, want %d\n", len(demo.textProto.GetFrames()), tc.wantFrames)
			}
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
			demo := NewDM2Parser()
			demo.binaryData = tc.data
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

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		inFile  string
		outFile string
	}{
		{
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
		},
		{
			name:    "fall and land with sound health highlight",
			inFile:  "/Users/joe/.quake2/baseq2/demos/test-dm1fall2.dm2",
			outFile: "/Users/joe/.quake2/baseq2/demos/test-dm1fall2-out.dm2",
		},
		{
			name:    "rocket shot with explosion",
			inFile:  "/Users/joe/.quake2/baseq2/demos/test-dm1rocket.dm2",
			outFile: "/Users/joe/.quake2/baseq2/demos/test-dm1rocket-out.dm2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewDM2Parser()
			content, err := os.ReadFile(tc.inFile)
			if err != nil {
				t.Error(err)
			}
			err = parser.Unmarshal(content)
			if err != nil {
				t.Error(err)
			}
			got, err := parser.Marshal()
			if err != nil {
				t.Error(err)
			}
			err = os.WriteFile(tc.outFile, got, 0644)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDemoDebug(t *testing.T) {
	/*
		tests := []struct {
			name   string
			inFile string
		}{

			{
				name:   "fall and land with sound health highlight",
				inFile: "/Users/joe/.quake2/baseq2/demos/test-dm1fall2.dm2",
			},

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
	*/
}
