package pak

import (
	"bytes"
	"os"
	"testing"

	pb "github.com/packetflinger/libq2/proto"
)

func TestUnmarshal(t *testing.T) {
	data, err := os.ReadFile("../testdata/test.pak")
	if err != nil {
		t.Error(err)
	}

	archive, err := Unmarshal(data)
	if err != nil {
		t.Error(err)
	}
	t.Error(archive)
}

func TestMarshal(t *testing.T) {
	data, err := os.ReadFile("../testdata/test.pak")
	if err != nil {
		t.Error(err)
	}

	archive, err := Unmarshal(data)
	if err != nil {
		t.Error(err)
	}

	data, err = Marshal(archive)
	if err != nil {
		t.Error(err)
	}
	err = os.WriteFile("../testdata/test-out.pak", data, 0777)
	if err != nil {
		t.Error(err)
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
