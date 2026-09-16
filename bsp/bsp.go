// Package bsp reads and writes Quake 2 .bsp map files, representing their
// contents as the protocol buffer messages defined in proto/v1/bsp.proto.
package bsp

import (
	"fmt"
	"os"
	"strings"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

const (
	// Magic is the 4-byte file identifier, "IBSP" read as a little-endian int32.
	Magic = int32(('P' << 24) + ('S' << 16) + ('B' << 8) + 'I')
	// Version is the only .bsp format version Quake 2 supports.
	Version = 38

	// LumpCount is how many lumps every .bsp file has.
	LumpCount = 19

	// HeaderSize is magic (4 bytes) + version (4 bytes) + the lump
	// directory (LumpCount * 8 bytes each).
	HeaderSize = 4 + 4 + LumpCount*8

	LumpEntities    = 0
	LumpPlanes      = 1
	LumpVertices    = 2
	LumpVisibility  = 3
	LumpNodes       = 4
	LumpTexInfo     = 5
	LumpFaces       = 6
	LumpLighting    = 7
	LumpLeaves      = 8
	LumpLeafFaces   = 9
	LumpLeafBrushes = 10
	LumpEdges       = 11
	LumpSurfEdges   = 12
	LumpModels      = 13
	LumpBrushes     = 14
	LumpBrushSides  = 15
	LumpPop         = 16
	LumpAreas       = 17
	LumpAreaPortals = 18
)

// lumpAlignment is the byte boundary each lump is padded to when writing,
// matching the alignment used by id Software's original map compilers.
const lumpAlignment = 4

// Open reads a .bsp file from disk and unmarshals it.
func Open(filename string) (*bpb.BSPFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	bsp, err := Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	tokens := strings.Split(filename, string(os.PathSeparator))
	name := strings.TrimSuffix(tokens[len(tokens)-1], ".bsp")
	bsp.Name = name
	bsp.Filename = filename
	return bsp, nil
}

// Save marshals a BSPFile and writes the resulting binary data to disk.
func Save(bsp *bpb.BSPFile, filename string) error {
	data, err := Marshal(bsp)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	return os.WriteFile(filename, data, 0644)
}

// parseHeader reads the magic number, version, and lump directory from the
// start of a .bsp file, returning the header proto and a function that
// slices out an individual lump's raw bytes by index.
func parseHeader(data []byte) (*bpb.BSPHeader, func(index int) ([]byte, error), error) {
	if len(data) < HeaderSize {
		return nil, nil, fmt.Errorf("file too short to contain a header: got %d bytes, want at least %d", len(data), HeaderSize)
	}

	hr := newReader(data[:HeaderSize])
	magic := hr.int32Val()
	if magic != Magic {
		return nil, nil, fmt.Errorf("bad magic number %#08x, want %#08x", uint32(magic), uint32(Magic))
	}
	version := hr.int32Val()

	header := &bpb.BSPHeader{Magic: magic, Version: version}
	var lumps [LumpCount]*bpb.BSPLump
	for i := 0; i < LumpCount; i++ {
		l := &bpb.BSPLump{Offset: hr.int32Val(), Length: hr.int32Val()}
		lumps[i] = l
		header.Lumps = append(header.Lumps, l)
	}

	lump := func(index int) ([]byte, error) {
		l := lumps[index]
		start, length := int(l.GetOffset()), int(l.GetLength())
		if start < 0 || length < 0 || start+length > len(data) {
			return nil, fmt.Errorf("lump %d out of bounds: offset=%d length=%d file=%d bytes", index, start, length, len(data))
		}
		return data[start : start+length], nil
	}

	return header, lump, nil
}

// EntityLumpBytes extracts the raw, unparsed bytes of a .bsp file's entity
// lump. This is what q2pro's CRC-16 checksum for hashed ".ent" override
// filenames is computed over (see EntChecksum): re-marshaling a parsed
// entity list back to text won't reproduce the same bytes, since entity
// property order isn't preserved once parsed into a map.
func EntityLumpBytes(data []byte) ([]byte, error) {
	_, lump, err := parseHeader(data)
	if err != nil {
		return nil, err
	}
	return lump(LumpEntities)
}

// Unmarshal parses the raw binary contents of a .bsp file into a BSPFile
// proto message.
func Unmarshal(data []byte) (*bpb.BSPFile, error) {
	header, lump, err := parseHeader(data)
	if err != nil {
		return nil, err
	}

	bsp := &bpb.BSPFile{Header: header}

	steps := []struct {
		index int
		name  string
		fn    func([]byte) error
	}{
		{LumpEntities, "entities", func(b []byte) (err error) { bsp.Entities, err = parseEntities(b); return }},
		{LumpPlanes, "planes", func(b []byte) (err error) { bsp.Planes, err = parsePlanes(b); return }},
		{LumpVertices, "vertices", func(b []byte) (err error) { bsp.Vertices, err = parseVertices(b); return }},
		{LumpVisibility, "visibility", func(b []byte) (err error) { bsp.Visibility, err = parseVisibility(b); return }},
		{LumpNodes, "nodes", func(b []byte) (err error) { bsp.Nodes, err = parseNodes(b); return }},
		{LumpTexInfo, "texture info", func(b []byte) (err error) { bsp.TextureInfo, err = parseTexInfo(b); return }},
		{LumpFaces, "faces", func(b []byte) (err error) { bsp.Faces, err = parseFaces(b); return }},
		{LumpLighting, "lighting", func(b []byte) error { bsp.Lighting = append([]byte{}, b...); return nil }},
		{LumpLeaves, "leaves", func(b []byte) (err error) { bsp.Leaves, err = parseLeaves(b); return }},
		{LumpLeafFaces, "leaf faces", func(b []byte) (err error) { bsp.LeafFaces, err = parseUint16Table(b); return }},
		{LumpLeafBrushes, "leaf brushes", func(b []byte) (err error) { bsp.LeafBrushes, err = parseUint16Table(b); return }},
		{LumpEdges, "edges", func(b []byte) (err error) { bsp.Edges, err = parseEdges(b); return }},
		{LumpSurfEdges, "surf edges", func(b []byte) (err error) { bsp.SurfEdges, err = parseSurfEdges(b); return }},
		{LumpModels, "models", func(b []byte) (err error) { bsp.Models, err = parseModels(b); return }},
		{LumpBrushes, "brushes", func(b []byte) (err error) { bsp.Brushes, err = parseBrushes(b); return }},
		{LumpBrushSides, "brush sides", func(b []byte) (err error) { bsp.BrushSides, err = parseBrushSides(b); return }},
		{LumpPop, "pop", func(b []byte) error { bsp.Pop = append([]byte{}, b...); return nil }},
		{LumpAreas, "areas", func(b []byte) (err error) { bsp.Areas, err = parseAreas(b); return }},
		{LumpAreaPortals, "area portals", func(b []byte) (err error) { bsp.AreaPortals, err = parseAreaPortals(b); return }},
	}

	for _, step := range steps {
		data, err := lump(step.index)
		if err != nil {
			return nil, err
		}
		if err := step.fn(data); err != nil {
			return nil, fmt.Errorf("parsing %s lump: %w", step.name, err)
		}
	}

	return bsp, nil
}

// Marshal serializes a BSPFile proto message back into the raw binary
// layout of a .bsp file, ready to be written to disk. The lumps are always
// written in a fixed order (entities, planes, vertices, ...) regardless of
// how they were laid out in whatever file the message originated from.
func Marshal(bsp *bpb.BSPFile) ([]byte, error) {
	lumps := [LumpCount][]byte{
		LumpEntities:    marshalEntities(bsp.GetEntities()),
		LumpVisibility:  marshalVisibility(bsp.GetVisibility()),
		LumpLighting:    bsp.GetLighting(),
		LumpLeafFaces:   marshalUint16Table(bsp.GetLeafFaces()),
		LumpLeafBrushes: marshalUint16Table(bsp.GetLeafBrushes()),
		LumpSurfEdges:   marshalSurfEdges(bsp.GetSurfEdges()),
		LumpPop:         bsp.GetPop(),
	}

	var err error
	if lumps[LumpPlanes], err = marshalPlanes(bsp.GetPlanes()); err != nil {
		return nil, fmt.Errorf("marshaling planes lump: %w", err)
	}
	lumps[LumpVertices] = marshalVertices(bsp.GetVertices())
	lumps[LumpNodes] = marshalNodes(bsp.GetNodes())
	if lumps[LumpTexInfo], err = marshalTexInfo(bsp.GetTextureInfo()); err != nil {
		return nil, fmt.Errorf("marshaling texture info lump: %w", err)
	}
	if lumps[LumpFaces], err = marshalFaces(bsp.GetFaces()); err != nil {
		return nil, fmt.Errorf("marshaling faces lump: %w", err)
	}
	lumps[LumpLeaves] = marshalLeaves(bsp.GetLeaves())
	lumps[LumpEdges] = marshalEdges(bsp.GetEdges())
	lumps[LumpModels] = marshalModels(bsp.GetModels())
	lumps[LumpBrushes] = marshalBrushes(bsp.GetBrushes())
	lumps[LumpBrushSides] = marshalBrushSides(bsp.GetBrushSides())
	lumps[LumpAreas] = marshalAreas(bsp.GetAreas())
	lumps[LumpAreaPortals] = marshalAreaPortals(bsp.GetAreaPortals())

	body := &writer{}
	var directory [LumpCount]*bpb.BSPLump
	for i, data := range lumps {
		body.pad(lumpAlignment)
		directory[i] = &bpb.BSPLump{Offset: int32(HeaderSize + body.buf.Len()), Length: int32(len(data))}
		body.bytes(data)
	}

	out := &writer{}
	out.int32Val(Magic)
	version := bsp.GetHeader().GetVersion()
	if version == 0 {
		version = Version
	}
	out.int32Val(version)
	for _, l := range directory {
		out.int32Val(l.GetOffset())
		out.int32Val(l.GetLength())
	}
	out.bytes(body.buf.Bytes())

	return out.buf.Bytes(), nil
}
