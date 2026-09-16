package bsp

import (
	"fmt"
	"os"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

const (
	// overrideMapNameSize matches Quake 2's MAX_QPATH, the fixed width of
	// the replacement map name field.
	overrideMapNameSize = 64

	// maxOverrideEntityString matches Quake 2's MAX_MAP_ENTSTRING. r1q2
	// rejects an override entity string that's zero or at/above this size.
	maxOverrideEntityString = 0x40000

	overrideFlagMapName  = 1 << 0
	overrideFlagChecksum = 1 << 1
	overrideFlagEntities = 1 << 2
)

// OpenOverride reads an r1q2 ".override" file from disk and unmarshals it.
func OpenOverride(filename string) (*bpb.BSPOverride, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	o, err := UnmarshalOverride(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	return o, nil
}

// SaveOverride marshals a BSPOverride and writes the resulting binary data
// to disk.
func SaveOverride(o *bpb.BSPOverride, filename string) error {
	data, err := MarshalOverride(o)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	return os.WriteFile(filename, data, 0644)
}

// UnmarshalOverride parses the raw binary contents of an r1q2 ".override"
// file. These are small sidecar files (conventionally named
// "<mapname>.bsp.override") that let a server transparently redirect
// clients to a different .bsp, report a specific checksum instead of
// computing one, and/or replace the entity string, all without modifying
// the .bsp itself.
//
// The format is a uint32 bitmask of which fields follow, then those
// fields in a fixed order: a 64-byte replacement map name (bit 0), a
// uint32 checksum (bit 1), and a uint32-length-prefixed entity string
// (bit 2). Each field is present only if its bit is set.
func UnmarshalOverride(data []byte) (*bpb.BSPOverride, error) {
	r := newReader(data)
	flags := r.uint32Val()
	if r.err != nil {
		return nil, fmt.Errorf("reading override flags: %w", r.err)
	}

	o := &bpb.BSPOverride{}

	if flags&overrideFlagMapName != 0 {
		name := r.fixedString(overrideMapNameSize)
		if r.err != nil {
			return nil, fmt.Errorf("reading override map name: %w", r.err)
		}
		o.MapName = &name
	}

	if flags&overrideFlagChecksum != 0 {
		checksum := r.uint32Val()
		if r.err != nil {
			return nil, fmt.Errorf("reading override checksum: %w", r.err)
		}
		o.Checksum = &checksum
	}

	if flags&overrideFlagEntities != 0 {
		length := r.uint32Val()
		if r.err != nil {
			return nil, fmt.Errorf("reading override entity string length: %w", r.err)
		}
		if length == 0 || length >= maxOverrideEntityString {
			return nil, fmt.Errorf("bad override entity string size %d, want 1-%d", length, maxOverrideEntityString-1)
		}
		entityText := r.take(int(length))
		if r.err != nil {
			return nil, fmt.Errorf("reading override entity string: %w", r.err)
		}
		entities, err := parseEntities(entityText)
		if err != nil {
			return nil, fmt.Errorf("parsing override entity string: %w", err)
		}
		o.Entities = entities
	}

	return o, nil
}

// MarshalOverride serializes a BSPOverride back into the raw binary layout
// of an r1q2 ".override" file, ready to be written to disk. Which fields
// are written is derived from which are set: MapName and Checksum are
// unset in proto3 optional fashion (nil means absent), and Entities is
// written only if non-empty.
func MarshalOverride(o *bpb.BSPOverride) ([]byte, error) {
	var flags uint32
	if o.MapName != nil {
		flags |= overrideFlagMapName
	}
	if o.Checksum != nil {
		flags |= overrideFlagChecksum
	}
	if len(o.GetEntities()) > 0 {
		flags |= overrideFlagEntities
	}

	w := &writer{}
	w.uint32Val(flags)

	if flags&overrideFlagMapName != 0 {
		if err := w.fixedString(o.GetMapName(), overrideMapNameSize); err != nil {
			return nil, fmt.Errorf("writing override map name: %w", err)
		}
	}

	if flags&overrideFlagChecksum != 0 {
		w.uint32Val(o.GetChecksum())
	}

	if flags&overrideFlagEntities != 0 {
		entityText := marshalEntities(o.GetEntities())
		if len(entityText) >= maxOverrideEntityString {
			return nil, fmt.Errorf("override entity string is %d bytes, want less than %d", len(entityText), maxOverrideEntityString)
		}
		w.uint32Val(uint32(len(entityText)))
		w.bytes(entityText)
	}

	return w.buf.Bytes(), nil
}
