package demo

import "testing"

func TestParseMVD2TextDemo(t *testing.T) {
	tests := []struct {
		name     string
		demofile string
	}{
		{
			name:     "Test1",
			demofile: "../testdata/temp.textmvd2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseMVD2TextDemo(tc.demofile)
			if err != nil {
				t.Error(err)
			}
			//t.Error()
		})
	}
}
