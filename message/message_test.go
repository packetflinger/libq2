package message

import (
	"testing"
)

func TestWriteLong(t *testing.T) {
	tests := []struct {
		desc      string
		input     int
		want      []byte
		wantmatch bool
	}{
		{
			desc:      "test_1",
			input:     -1,
			want:      []byte{255, 255, 255, 255},
			wantmatch: true,
		},
		{
			desc:      "test_2",
			input:     0,
			want:      []byte{0, 0, 0, 0},
			wantmatch: true,
		},
		{
			desc:      "test_3",
			input:     3453445,
			want:      []byte{5, 178, 52, 0},
			wantmatch: true,
		},
	}

	for _, test := range tests {
		msg := MessageBuffer{}
		msg.WriteLong(int32(test.input))
		for i := range msg.Buffer {
			if msg.Buffer[i] != test.want[i] {
				t.Errorf("%s failed\n", test.desc)
			}
		}
	}
}
