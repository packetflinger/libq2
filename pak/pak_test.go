package pak

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/packetflinger/libq2/proto"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string // bytes encoded as hex
		want    *pb.PAKArchive
		wantErr bool
	}{
		{
			name:    "short pak",
			input:   "1234c0ffeeface4321",
			wantErr: true,
		},
		{
			name:    "Garbage pak",
			input:   "1234c0ffeeface4321001234c0ffeeface4321001234c0ffeeface4321001234c0ffeeface4321001234c0ffeeface4321001234c0ffeeface4321001234c0ffeeface4321001234c0ffeeface432100",
			wantErr: true,
		},
		{
			name:  "Valid pak",
			input: "5041434b1500000080000000010203040506070809746573742e6366670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0000000500000074657374322e63666700000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001100000004000000",
			want: &pb.PAKArchive{
				Files: []*pb.PAKFile{
					{
						Name: "test.cfg",
						Data: []byte{1, 2, 3, 4, 5},
					},
					{
						Name: "test2.cfg",
						Data: []byte{6, 7, 8, 9},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Fatalf("unable to decode hex string %q: %v", tc.input, err)
			}
			got, err := Unmarshal(in)
			if !tc.wantErr && (err != nil) {
				t.Error()
			} else {
				ignore := protocmp.IgnoreFields(&pb.PAKFile{}, "location", "length", "hash")
				diff := cmp.Diff(got, tc.want, protocmp.Transform(), ignore)
				if diff != "" {
					t.Error("\ngot:\n", prototext.Format(got), "\nwant\n", prototext.Format(tc.want))
				}
			}
		})
	}
}

func TestExtractFile(t *testing.T) {
	tests := []struct {
		name string
		pak  *pb.PAKArchive
		file string
		want []byte
	}{
		{
			name: "Test 1",
			pak: &pb.PAKArchive{
				Files: []*pb.PAKFile{
					{
						Name: "test.cfg",
						Data: []byte{1, 2, 3, 4, 5},
					},
					{
						Name: "test2.cfg",
						Data: []byte{6, 7, 8, 9},
					},
				},
			},
			file: "test2.cfg",
			want: []byte{6, 7, 8, 9},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ExtractFile(tc.pak, tc.file)
			if err != nil {
				t.Error()
			}
			if !bytes.Equal(got.Data, tc.want) {
				t.Error("\ngot:\n", got.Data, "\nwant\n", tc.want)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		pak     *pb.PAKArchive
		want    string // bytes encoded as hex
		wantErr bool
	}{
		{
			name: "Valid pak",
			pak: &pb.PAKArchive{
				Files: []*pb.PAKFile{
					{
						Name: "test.cfg",
						Data: []byte{1, 2, 3, 4, 5},
					},
					{
						Name: "test2.cfg",
						Data: []byte{6, 7, 8, 9},
					},
				},
			},
			want:    "5041434b1500000080000000010203040506070809746573742e6366670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0000000500000074657374322e63666700000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001100000004000000",
			wantErr: false,
		},
		{
			name: "duplicate names",
			pak: &pb.PAKArchive{
				Files: []*pb.PAKFile{
					{
						Name: "test.cfg",
						Data: []byte{1, 2, 3, 4, 5},
					},
					{
						Name: "test.cfg",
						Data: []byte{6, 7, 8, 9},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Marshal(tc.pak)
			if !tc.wantErr && (err != nil) {
				t.Error()
			} else {
				wantBytes, err := hex.DecodeString(tc.want)
				if err != nil {
					t.Fatalf("unable to decode want %q", tc.want)
				}
				if !bytes.Equal(got, wantBytes) {
					t.Error("\ngot:\n", hex.EncodeToString(got), "\nwant\n", tc.want)
				}
			}
		})
	}
}

func TestDataHash(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want string
	}{
		{
			name: "test1",
			in:   []byte("test1"),
			want: "1b4f0e9851971998e732078544c96b36c3d01cedf7caa332359d6f1d83567014",
		},
		{
			name: "test2",
			in:   []byte("test2"),
			want: "60303ae22b998861bce3b28f33eec1be758a213c86c93c076dbe9f558c11c752",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dataHash(tc.in)
			if got != tc.want {
				t.Errorf("dataHash(%v) = %s, want %s\n", tc.in, got, tc.want)
			}
		})
	}
}
