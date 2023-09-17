package message

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
	msg1.Append(msg2)

	msg1.Index = 0
	if got := msg1.ReadShort(); got != 257 {
		t.Errorf("Append failed - got %d, want 257\n", got)
	}
}

func TestWriteByte(t *testing.T) {
	tests := []struct {
		desc  string
		input int
		want  []byte
	}{
		{
			desc:  "TEST 1",
			input: 5,
			want:  []byte{5},
		},
		{
			desc:  "TEST 2",
			input: 255,
			want:  []byte{255},
		},
		{
			desc:  "TEST 3",
			input: 100,
			want:  []byte{100},
		},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			msg := MessageBuffer{}
			msg.WriteByte(byte(tc.input))
			got := msg.Buffer
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Error("got diff:", diff)
			}
		})
	}
}

func TestWriteShort(t *testing.T) {
	tests := []struct {
		desc  string
		input uint16
		want  []uint8
	}{
		{
			desc:  "TEST 1",
			input: 5,
			want:  []uint8{5, 0},
		},
		{
			desc:  "TEST 2",
			input: 65535,
			want:  []uint8{255, 255},
		},
		{
			desc:  "TEST 3",
			input: 257,
			want:  []uint8{1, 1},
		},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			msg := MessageBuffer{}
			msg.WriteShort(tc.input)
			got := msg.Buffer
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Error("got diff:", diff)
			}
		})
	}
}

func TestWriteString(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  []uint8
	}{
		{
			desc:  "TEST 1",
			input: "q2dm1",
			want:  []uint8{0x71, 0x32, 0x64, 0x6d, 0x31, 0x00}, // always ends in null
		},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			msg := MessageBuffer{}
			msg.WriteString(tc.input)
			got := msg.Buffer
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Error("got diff:", diff)
			}
		})
	}
}

func TestReadULong(t *testing.T) {
	tests := []struct {
		desc  string
		input MessageBuffer
		want  uint32
	}{
		{
			desc:  "test1",
			input: NewMessageBuffer([]byte{255, 255, 255, 255}),
			want:  4294967295,
		},
		{
			desc:  "test2",
			input: NewMessageBuffer([]byte{128, 0, 0, 0}),
			want:  128,
		},
		{
			desc:  "test3",
			input: NewMessageBuffer([]byte{0, 0, 0, 128}),
			want:  2147483648,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.input.ReadULong()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestReadLong(t *testing.T) {
	tests := []struct {
		desc  string
		input MessageBuffer
		want  int32
	}{
		{
			desc:  "test1",
			input: NewMessageBuffer([]byte{255, 255, 255, 255}),
			want:  -1,
		},
		{
			desc:  "test2",
			input: NewMessageBuffer([]byte{128, 0, 0, 0}),
			want:  128,
		},
		{
			desc:  "test3",
			input: NewMessageBuffer([]byte{0, 0, 0, 128}),
			want:  -2147483648,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.input.ReadLong()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}
