package bsp

import (
	"bytes"
	"fmt"
	"os"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// UnmarshalEnt parses the raw contents of a q2pro ".ent" override file.
// Unlike r1q2's ".override" format (see UnmarshalOverride), a ".ent" file
// has no header at all: its entire contents are the replacement entity
// string, in the same brace-delimited text format as a .bsp's own entity
// lump.
func UnmarshalEnt(data []byte) ([]*bpb.BSPEntity, error) {
	return parseEntities(data)
}

// MarshalEnt encodes entities into the raw contents of a q2pro ".ent"
// file: just the brace-delimited entity text. No trailing NUL is written;
// q2pro NUL-terminates whatever it reads from disk itself, so one isn't
// needed, and omitting it keeps the file plain, human-editable text.
func MarshalEnt(entities []*bpb.BSPEntity) []byte {
	return bytes.TrimSuffix(marshalEntities(entities), []byte{0})
}

// OpenEnt reads a q2pro ".ent" file from disk and unmarshals it.
func OpenEnt(filename string) ([]*bpb.BSPEntity, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	entities, err := UnmarshalEnt(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	return entities, nil
}

// SaveEnt marshals entities and writes the resulting q2pro ".ent" file to
// disk.
//
// Since r1q2's BSPOverride and q2pro's .ent both ultimately just carry a
// []*bpb.BSPEntity, converting between the two formats needs no dedicated
// function - read one and save the other's Entities:
//
//	o, err := bsp.OpenOverride("maps/q2dm1.bsp.override")
//	...
//	err = bsp.SaveEnt(o.GetEntities(), "maps/q2dm1.ent")
//
// Note that only the entity replacement travels across: BSPOverride's
// MapName/Checksum redirects have no equivalent in the plain .ent format.
func SaveEnt(entities []*bpb.BSPEntity, filename string) error {
	return os.WriteFile(filename, MarshalEnt(entities), 0644)
}

// EntChecksum computes the CRC-16/CCITT-FALSE checksum q2pro uses to name
// hashed ".ent" override files (see HashedEntFileName). rawEntityLump must
// be a .bsp file's raw, unparsed entity lump bytes (see EntityLumpBytes) -
// not a re-marshaled entity list - since the checksum is sensitive to the
// exact original bytes, and property order isn't preserved once entities
// are parsed into a map.
func EntChecksum(rawEntityLump []byte) uint16 {
	// q2pro excludes the entity lump's trailing NUL from the checksum.
	if n := len(rawEntityLump); n > 0 && rawEntityLump[n-1] == 0 {
		rawEntityLump = rawEntityLump[:n-1]
	}
	return crc16CCITT(rawEntityLump)
}

// EntFileName returns the plain (unhashed) q2pro override filename for a
// map name, ex: EntFileName("q2dm1") == "q2dm1.ent".
func EntFileName(mapName string) string {
	return mapName + ".ent"
}

// HashedEntFileName returns the hash-qualified q2pro override filename for
// a map, given that map's original .bsp entity lump bytes. q2pro tries
// this form first and only serves it while the target .bsp's entity lump
// still matches the embedded hash, falling back to the plain EntFileName
// otherwise. See EntChecksum.
func HashedEntFileName(mapName string, rawEntityLump []byte) string {
	return fmt.Sprintf("%s@%04x.ent", mapName, EntChecksum(rawEntityLump))
}

// crc16CCITT implements the same CRC-16/CCITT-FALSE variant (polynomial
// 0x1021, initial value 0xffff, no reflection, no final XOR) as q2pro's
// CRC_Block.
func crc16CCITT(data []byte) uint16 {
	crc := uint16(0xffff)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
