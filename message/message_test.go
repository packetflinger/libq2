package message

import (
	"encoding/hex"
	"strings"
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
		msg := Buffer{}
		msg.WriteLong(test.input)
		for i := range msg.Data {
			if msg.Data[i] != test.want[i] {
				t.Errorf("%s failed\n", test.desc)
			}
		}
	}
}

func TestAppend(t *testing.T) {
	msg1 := Buffer{}
	msg2 := Buffer{}

	msg1.WriteByte(1)
	msg2.WriteByte(1)
	msg1.Append(msg2)

	msg1.Index = 0
	if got := msg1.ReadShort(); got != 257 {
		t.Errorf("Append failed - got %d, want 257\n", got)
	}
}

func TestReadByte(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "test1",
			input: "ff",
			want:  255,
		},
		{
			name:  "test2",
			input: "80",
			want:  128,
		},
		{
			name:  "test3",
			input: "bb",
			want:  187,
		},
		{
			name:  "test4",
			input: "15",
			want:  21,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			in := NewBuffer(bytes)
			got := in.ReadByte()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
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
			msg := Buffer{}
			msg.WriteByte(int(tc.input))
			got := msg.Data
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Error("got diff:", diff)
			}
		})
	}
}

func TestReadChar(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "test1",
			input: "ff",
			want:  -1,
		},
		{
			name:  "test2",
			input: "80",
			want:  -128,
		},
		{
			name:  "test3",
			input: "bb",
			want:  -69,
		},
		{
			name:  "test4",
			input: "15",
			want:  21,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			in := NewBuffer(bytes)
			got := in.ReadChar()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestWriteChar(t *testing.T) {
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
			input: -1,
			want:  []byte{255},
		},
		{
			desc:  "TEST 3",
			input: -6,
			want:  []byte{250},
		},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			msg := Buffer{}
			msg.WriteByte(int(tc.input))
			got := msg.Data
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Error("got diff:", diff)
			}
		})
	}
}

func TestReadShort(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  int
	}{
		{
			desc:  "test1",
			input: "ffff",
			want:  -1,
		},
		{
			desc:  "test2",
			input: "8000",
			want:  128,
		},
		{
			desc:  "test3",
			input: "0080",
			want:  -32768,
		},
		{
			desc:  "test4",
			input: "ff7f",
			want:  32767,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			in := NewBuffer(bytes)
			got := in.ReadShort()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestWriteShort(t *testing.T) {
	tests := []struct {
		desc  string
		input int
		want  []byte
	}{
		{
			desc:  "test1",
			input: 5,
			want:  []byte{5, 0},
		},
		{
			desc:  "test2",
			input: -1,
			want:  []byte{255, 255},
		},
		{
			desc:  "test3",
			input: 257,
			want:  []byte{1, 1},
		},
		{
			desc:  "test4",
			input: 65538,
			want:  []byte{2, 0},
		},
		{
			desc:  "test5",
			input: 65535,
			want:  []byte{255, 255},
		},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			msg := Buffer{}
			msg.WriteShort(tc.input)
			got := msg.Data
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Error("got diff:", diff)
			}
		})
	}
}

func TestReadWord(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  int
	}{
		{
			desc:  "test1",
			input: "ffff",
			want:  65535,
		},
		{
			desc:  "test2",
			input: "8000",
			want:  128,
		},
		{
			desc:  "test3",
			input: "0080",
			want:  32768,
		},
		{
			desc:  "test4",
			input: "ff7f",
			want:  32767,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			in := NewBuffer(bytes)
			got := in.ReadWord()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestWriteWord(t *testing.T) {
	tests := []struct {
		desc  string
		input int
		want  []byte
	}{
		{
			desc:  "test1",
			input: 5,
			want:  []byte{5, 0},
		},
		{
			desc:  "test2",
			input: 65535,
			want:  []byte{255, 255},
		},
		{
			desc:  "test3",
			input: 257,
			want:  []byte{1, 1},
		},
		{
			desc:  "test4",
			input: 65538,
			want:  []byte{2, 0},
		},
		{
			desc:  "test5",
			input: -1,
			want:  []byte{255, 255},
		},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			msg := Buffer{}
			msg.WriteWord(tc.input)
			got := msg.Data
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
			msg := Buffer{}
			msg.WriteString(tc.input)
			got := msg.Data
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Error("got diff:", diff)
			}
		})
	}
}

func TestReadLong(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  int
	}{
		{
			desc:  "test1",
			input: "ffffffff",
			want:  -1,
		},
		{
			desc:  "test2",
			input: "80000000",
			want:  128,
		},
		{
			desc:  "test3",
			input: "00000080",
			want:  -2147483648,
		},
		{
			desc:  "test4",
			input: "ffffff7f",
			want:  2147483647,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			in := NewBuffer(bytes)
			got := in.ReadLong()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestReadString(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		start int // start position
		want  string
	}{
		{
			desc:  "empty input",
			input: "",
			start: 0,
			want:  "",
		},
		{
			desc:  "null terminated",
			input: "6a75 7374 2061 696d 2064 6f77 6e00",
			start: 0,
			want:  "just aim down",
		},
		{
			desc:  "no ending null",
			input: "6a75 7374 2061 696d 2064 6f77 6e",
			start: 0,
			want:  "just aim down",
		},
		{
			desc:  "mid null",
			input: "6a75 7374 2061 696d 0000 2064 6f77 6e00",
			start: 0,
			want:  "just aim",
		},
		{
			desc:  "beginning null",
			input: "0000 6a75 7374 2061 696d 2064 6f77 6e00",
			start: 0,
			want:  "",
		},
		{
			desc:  "at end",
			input: "0000 6a75 7374 2061 696d 2064 6f77 6e00",
			start: 15,
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bytes, err := hex.DecodeString(strings.ReplaceAll(tc.input, " ", ""))
			if err != nil {
				t.Error(err)
			}
			in := NewBuffer(bytes)
			in.Index = tc.start
			got := in.ReadString()
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}
