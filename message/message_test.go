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

func TestAppend(t *testing.T) {
	msg1 := MessageBuffer{}
	msg2 := MessageBuffer{}

	msg1.WriteByte(1)
	msg2.WriteByte(1)
	msg1.Append(&msg2)

	msg1.Index = 0
	if got := msg1.ReadShort(); got != 257 {
		t.Errorf("Append failed - got %d, want 257\n", got)
	}
}
