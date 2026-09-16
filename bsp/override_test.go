package bsp

import (
	"bytes"
	"encoding/binary"
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
	"google.golang.org/protobuf/proto"
)

// buildOverride hand-encodes an .override file the way r1q2's CM_LoadMap
// reads it, independent of our own Marshal/Unmarshal code, so tests can
// check we actually match the real wire format and not just ourselves.
func buildOverride(t *testing.T, mapName *string, checksum *uint32, entityString []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	var flags uint32
	if mapName != nil {
		flags |= overrideFlagMapName
	}
	if checksum != nil {
		flags |= overrideFlagChecksum
	}
	if entityString != nil {
		flags |= overrideFlagEntities
	}
	binary.Write(&buf, binary.LittleEndian, flags)
	if mapName != nil {
		name := make([]byte, overrideMapNameSize)
		copy(name, *mapName)
		buf.Write(name)
	}
	if checksum != nil {
		binary.Write(&buf, binary.LittleEndian, *checksum)
	}
	if entityString != nil {
		binary.Write(&buf, binary.LittleEndian, uint32(len(entityString)))
		buf.Write(entityString)
	}
	return buf.Bytes()
}

func TestUnmarshalOverrideAllFields(t *testing.T) {
	name := "maps/other.bsp"
	checksum := uint32(0xdeadbeef)
	entityText := []byte("{\n\"classname\" \"worldspawn\"\n}\n\x00")

	data := buildOverride(t, &name, &checksum, entityText)

	got, err := UnmarshalOverride(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetMapName() != name {
		t.Errorf("map name - have: %q, want: %q", got.GetMapName(), name)
	}
	if got.GetChecksum() != checksum {
		t.Errorf("checksum - have: %#x, want: %#x", got.GetChecksum(), checksum)
	}
	if len(got.GetEntities()) != 1 {
		t.Fatalf("entity count - have: %d, want: 1", len(got.GetEntities()))
	}
	if got.GetEntities()[0].GetClassName() != "worldspawn" {
		t.Errorf("entity classname - have: %q, want: %q", got.GetEntities()[0].GetClassName(), "worldspawn")
	}
}

func TestUnmarshalOverrideNoFields(t *testing.T) {
	data := buildOverride(t, nil, nil, nil)
	if len(data) != 4 {
		t.Fatalf("expected a bare 4-byte flags field, got %d bytes", len(data))
	}
	got, err := UnmarshalOverride(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.MapName != nil {
		t.Errorf("map name - have: %v, want: nil", got.MapName)
	}
	if got.Checksum != nil {
		t.Errorf("checksum - have: %v, want: nil", got.Checksum)
	}
	if len(got.GetEntities()) != 0 {
		t.Errorf("entities - have: %d, want: 0", len(got.GetEntities()))
	}
}

func TestUnmarshalOverrideMapNameOnly(t *testing.T) {
	name := "maps/q2dm1.bsp"
	data := buildOverride(t, &name, nil, nil)
	got, err := UnmarshalOverride(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetMapName() != name {
		t.Errorf("map name - have: %q, want: %q", got.GetMapName(), name)
	}
	if got.Checksum != nil {
		t.Errorf("checksum - have: %v, want: nil", got.Checksum)
	}
}

func TestUnmarshalOverrideBadEntityLength(t *testing.T) {
	// length field of 0 is explicitly rejected by r1q2.
	data := buildOverride(t, nil, nil, []byte{})
	if _, err := UnmarshalOverride(data); err == nil {
		t.Error("expected an error for a zero-length entity string")
	}
}

func TestUnmarshalOverrideEntityLengthOverrunsBuffer(t *testing.T) {
	data := buildOverride(t, nil, nil, []byte("{}\x00"))
	truncated := data[:len(data)-1]
	if _, err := UnmarshalOverride(truncated); err == nil {
		t.Error("expected an error when the entity string is shorter than its length prefix claims")
	}
}

func TestOverrideRoundTrip(t *testing.T) {
	name := "maps/q2dm1.bsp"
	checksum := uint32(12345)
	want := &bpb.BSPOverride{
		MapName:  &name,
		Checksum: &checksum,
		Entities: []*bpb.BSPEntity{
			{ClassName: "worldspawn", Properties: map[string]string{"classname": "worldspawn"}},
			{ClassName: "info_player_start", Properties: map[string]string{"classname": "info_player_start", "origin": "0 0 0"}},
		},
	}

	data, err := MarshalOverride(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalOverride(data)
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(want, got) {
		t.Errorf("round trip mismatch:\nwant: %v\nhave: %v", want, got)
	}
}

func TestOverrideRoundTripEmpty(t *testing.T) {
	want := &bpb.BSPOverride{}
	data, err := MarshalOverride(want)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 4 {
		t.Fatalf("expected a bare 4-byte flags field for an empty override, got %d bytes", len(data))
	}
	got, err := UnmarshalOverride(data)
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(want, got) {
		t.Errorf("round trip mismatch:\nwant: %v\nhave: %v", want, got)
	}
}

func TestMarshalOverrideMapNameTooLong(t *testing.T) {
	name := "maps/" + string(make([]byte, overrideMapNameSize)) // way over 64 bytes
	_, err := MarshalOverride(&bpb.BSPOverride{MapName: &name})
	if err == nil {
		t.Error("expected an error for a map name that doesn't fit in the fixed field")
	}
}

func TestMarshalOverrideOnlyChecksum(t *testing.T) {
	checksum := uint32(999)
	data, err := MarshalOverride(&bpb.BSPOverride{Checksum: &checksum})
	if err != nil {
		t.Fatal(err)
	}
	// flags (4) + checksum (4), no map name or entity string.
	if len(data) != 8 {
		t.Fatalf("marshaled size - have: %d, want: 8", len(data))
	}
	got, err := UnmarshalOverride(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetChecksum() != checksum {
		t.Errorf("checksum - have: %d, want: %d", got.GetChecksum(), checksum)
	}
	if got.MapName != nil {
		t.Errorf("map name - have: %v, want: nil", got.MapName)
	}
}
