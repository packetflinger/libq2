package bsp

import (
	"os"
	"path/filepath"
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestEntRoundTrip(t *testing.T) {
	want := []*bpb.BSPEntity{
		{ClassName: "worldspawn", Properties: map[string]string{"classname": "worldspawn", "message": "test"}},
		{ClassName: "info_player_start", Properties: map[string]string{"classname": "info_player_start", "origin": "1 2 3"}},
	}

	data := MarshalEnt(want)
	if len(data) == 0 || data[len(data)-1] == 0 {
		t.Fatalf("expected plain text with no trailing NUL, got: %q", data)
	}

	got, err := UnmarshalEnt(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("entity count - have: %d, want: %d", len(got), len(want))
	}
	for i := range want {
		if got[i].GetClassName() != want[i].GetClassName() {
			t.Errorf("entity %d classname - have: %q, want: %q", i, got[i].GetClassName(), want[i].GetClassName())
		}
		for k, v := range want[i].GetProperties() {
			if got[i].GetProperties()[k] != v {
				t.Errorf("entity %d property %q - have: %q, want: %q", i, k, got[i].GetProperties()[k], v)
			}
		}
	}
}

func TestOpenSaveEnt(t *testing.T) {
	want := []*bpb.BSPEntity{
		{ClassName: "worldspawn", Properties: map[string]string{"classname": "worldspawn"}},
	}
	path := filepath.Join(t.TempDir(), "q2dm1.ent")

	if err := SaveEnt(want, path); err != nil {
		t.Fatal(err)
	}
	got, err := OpenEnt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].GetClassName() != "worldspawn" {
		t.Errorf("round trip via disk - have: %v, want classname worldspawn", got)
	}

	// Confirm the file on disk is plain text, not the r1q2 binary format.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw[:1]) != "{" {
		t.Errorf("expected a .ent file to start with a raw '{', got: %q", raw[:1])
	}
}

// TestOverrideToEnt is the scenario the format was added for: read an
// r1q2 ".override" file and save its entity replacement as a q2pro ".ent"
// file, with no other conversion step required since both formats share
// the same []*bpb.BSPEntity representation.
func TestOverrideToEnt(t *testing.T) {
	overridePath := filepath.Join(t.TempDir(), "q2dm1.bsp.override")
	entPath := filepath.Join(t.TempDir(), "q2dm1.ent")

	original := []*bpb.BSPEntity{
		{ClassName: "worldspawn", Properties: map[string]string{"classname": "worldspawn"}},
		{ClassName: "info_player_start", Properties: map[string]string{"classname": "info_player_start", "origin": "64 312 408"}},
	}
	if err := SaveOverride(&bpb.BSPOverride{Entities: original}, overridePath); err != nil {
		t.Fatal(err)
	}

	o, err := OpenOverride(overridePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveEnt(o.GetEntities(), entPath); err != nil {
		t.Fatal(err)
	}

	got, err := OpenEnt(entPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(original) {
		t.Fatalf("entity count - have: %d, want: %d", len(got), len(original))
	}
	for i := range original {
		if got[i].GetClassName() != original[i].GetClassName() {
			t.Errorf("entity %d classname - have: %q, want: %q", i, got[i].GetClassName(), original[i].GetClassName())
		}
	}
}

func TestEntChecksum(t *testing.T) {
	// Standard CRC-16/CCITT-FALSE check value; also cross-checked directly
	// against q2pro's own crctable in crc.c.
	if got, want := EntChecksum([]byte("123456789")), uint16(0x29b1); got != want {
		t.Errorf("checksum - have: %#04x, want: %#04x", got, want)
	}
	// A trailing NUL (as a real .bsp entity lump always has) must be
	// excluded from the checksum, matching CRC_Block(str, numchars-1).
	if got, want := EntChecksum([]byte("123456789\x00")), uint16(0x29b1); got != want {
		t.Errorf("checksum with trailing NUL - have: %#04x, want: %#04x", got, want)
	}
}

func TestEntFileNames(t *testing.T) {
	if got, want := EntFileName("q2dm1"), "q2dm1.ent"; got != want {
		t.Errorf("EntFileName - have: %q, want: %q", got, want)
	}
	got := HashedEntFileName("q2dm1", []byte("123456789"))
	if want := "q2dm1@29b1.ent"; got != want {
		t.Errorf("HashedEntFileName - have: %q, want: %q", got, want)
	}
}

func TestEntityLumpBytesMatchesParsedEntities(t *testing.T) {
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := EntityLumpBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 4537 {
		t.Fatalf("raw entity lump length - have: %d, want: 4537", len(raw))
	}

	entities, err := UnmarshalEnt(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 65 {
		t.Errorf("entity count from raw lump - have: %d, want: 65", len(entities))
	}

	// Sanity check the checksum/filename helpers work end to end against a
	// real map's raw entity lump.
	name := HashedEntFileName("backup", raw)
	if filepath.Ext(name) != ".ent" {
		t.Errorf("HashedEntFileName - have: %q, want a .ent file", name)
	}
}
