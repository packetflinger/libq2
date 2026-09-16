package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// leafSize is the on-disk size of a dleaf_t: contents (4 bytes), cluster +
// area (2 bytes each), a short mins/maxs bounding box (12 bytes), and 4
// unsigned short table indices (2 bytes each).
const leafSize = 28

// tableEntrySize is the width of one entry in the leaf-faces and
// leaf-brushes index tables.
const tableEntrySize = 2

func parseLeaves(data []byte) ([]*bpb.BSPLeaf, error) {
	if len(data)%leafSize != 0 {
		return nil, fmt.Errorf("leaves lump size %d is not a multiple of %d", len(data), leafSize)
	}
	count := len(data) / leafSize
	leaves := make([]*bpb.BSPLeaf, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		leaves = append(leaves, &bpb.BSPLeaf{
			Contents:       r.int32Val(),
			Cluster:        r.int16(),
			Area:           r.int16(),
			Bounds:         r.shortBounds(),
			FirstLeafFace:  r.uint16(),
			NumLeafFaces:   r.uint16(),
			FirstLeafBrush: r.uint16(),
			NumLeafBrushes: r.uint16(),
		})
	}
	return leaves, r.err
}

func marshalLeaves(leaves []*bpb.BSPLeaf) []byte {
	w := &writer{}
	for _, l := range leaves {
		w.int32Val(l.GetContents())
		w.int16(l.GetCluster())
		w.int16(l.GetArea())
		w.shortBounds(l.GetBounds())
		w.uint16(l.GetFirstLeafFace())
		w.uint16(l.GetNumLeafFaces())
		w.uint16(l.GetFirstLeafBrush())
		w.uint16(l.GetNumLeafBrushes())
	}
	return w.buf.Bytes()
}

// parseUint16Table decodes the leaf-faces and leaf-brushes lumps, both of
// which are flat arrays of unsigned 16-bit indices.
func parseUint16Table(data []byte) ([]uint32, error) {
	if len(data)%tableEntrySize != 0 {
		return nil, fmt.Errorf("index table size %d is not a multiple of %d", len(data), tableEntrySize)
	}
	count := len(data) / tableEntrySize
	table := make([]uint32, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		table = append(table, r.uint16())
	}
	return table, r.err
}

func marshalUint16Table(table []uint32) []byte {
	w := &writer{}
	for _, v := range table {
		w.uint16(v)
	}
	return w.buf.Bytes()
}
