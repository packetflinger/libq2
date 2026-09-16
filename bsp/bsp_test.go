package bsp

import (
	"os"
	"testing"

	"google.golang.org/protobuf/proto"
)

const testFile = "../testdata/backup.bsp"

func TestOpen(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if bsp.GetName() != "backup" {
		t.Errorf("Name - have: %q, want: %q", bsp.GetName(), "backup")
	}
	if bsp.GetFilename() != testFile {
		t.Errorf("Filename - have: %q, want: %q", bsp.GetFilename(), testFile)
	}
	if bsp.GetHeader().GetMagic() != Magic {
		t.Errorf("Magic - have: %#08x, want: %#08x", uint32(bsp.GetHeader().GetMagic()), uint32(Magic))
	}
	if bsp.GetHeader().GetVersion() != Version {
		t.Errorf("Version - have: %d, want: %d", bsp.GetHeader().GetVersion(), Version)
	}
	if len(bsp.GetHeader().GetLumps()) != LumpCount {
		t.Errorf("Lump directory - have: %d entries, want: %d", len(bsp.GetHeader().GetLumps()), LumpCount)
	}
	if got, want := len(bsp.GetLighting()), 91431; got != want {
		t.Errorf("Lighting - have: %d bytes, want: %d", got, want)
	}
	if got, want := len(bsp.GetPop()), 256; got != want {
		t.Errorf("Pop - have: %d bytes, want: %d", got, want)
	}
}

func TestUnmarshalBadMagic(t *testing.T) {
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	corrupt := append([]byte{}, data...)
	corrupt[0] ^= 0xff
	if _, err := Unmarshal(corrupt); err == nil {
		t.Error("expected an error for a bad magic number, got nil")
	}
}

func TestUnmarshalTooShort(t *testing.T) {
	if _, err := Unmarshal([]byte{1, 2, 3}); err == nil {
		t.Error("expected an error for a too-short file, got nil")
	}
}

// Round-tripping through Marshal/Unmarshal should produce a file that
// decodes back into an equivalent proto message. The raw bytes aren't
// expected to match exactly (lump ordering/padding and entity property
// order aren't preserved), but the structured data should be.
func TestRoundTrip(t *testing.T) {
	original, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}

	data, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	roundTripped, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("re-parsing marshaled data: %v", err)
	}

	// Name/Filename aren't part of the on-disk format, so they don't
	// survive a Marshal/Unmarshal cycle. The lump directory's offsets also
	// won't match: Marshal always lays lumps out in a fixed order, which
	// generally differs from whatever order the original map compiler
	// used. Neither is a meaningful part of the map data, so clear them
	// before comparing everything else.
	original.Name = ""
	original.Filename = ""
	if got, want := roundTripped.GetHeader().GetMagic(), original.GetHeader().GetMagic(); got != want {
		t.Errorf("magic - have: %#08x, want: %#08x", uint32(got), uint32(want))
	}
	if got, want := roundTripped.GetHeader().GetVersion(), original.GetHeader().GetVersion(); got != want {
		t.Errorf("version - have: %d, want: %d", got, want)
	}
	original.Header = nil
	roundTripped.Header = nil

	if !proto.Equal(original, roundTripped) {
		t.Error("round-tripped BSPFile does not match the original")
	}
}

func TestMarshalWritesValidHeader(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	data, err := Marshal(bsp)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < HeaderSize {
		t.Fatalf("marshaled data too short: %d bytes", len(data))
	}
	if _, err := Unmarshal(data); err != nil {
		t.Errorf("marshaled data did not parse back: %v", err)
	}
}
