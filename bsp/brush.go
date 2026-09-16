package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// brushSize is the on-disk size of a dbrush_t: firstside + numsides +
// contents (4 bytes each).
const brushSize = 12

// brushSideSize is the on-disk size of a dbrushside_t: an unsigned short
// plane index and a signed short texinfo index.
const brushSideSize = 4

func parseBrushes(data []byte) ([]*bpb.BSPBrush, error) {
	if len(data)%brushSize != 0 {
		return nil, fmt.Errorf("brushes lump size %d is not a multiple of %d", len(data), brushSize)
	}
	count := len(data) / brushSize
	brushes := make([]*bpb.BSPBrush, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		brushes = append(brushes, &bpb.BSPBrush{
			FirstSide: r.int32Val(),
			NumSides:  r.int32Val(),
			Contents:  r.int32Val(),
		})
	}
	return brushes, r.err
}

func marshalBrushes(brushes []*bpb.BSPBrush) []byte {
	w := &writer{}
	for _, b := range brushes {
		w.int32Val(b.GetFirstSide())
		w.int32Val(b.GetNumSides())
		w.int32Val(b.GetContents())
	}
	return w.buf.Bytes()
}

func parseBrushSides(data []byte) ([]*bpb.BSPBrushSide, error) {
	if len(data)%brushSideSize != 0 {
		return nil, fmt.Errorf("brush sides lump size %d is not a multiple of %d", len(data), brushSideSize)
	}
	count := len(data) / brushSideSize
	sides := make([]*bpb.BSPBrushSide, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		sides = append(sides, &bpb.BSPBrushSide{
			PlaneNum: r.uint16(),
			Texinfo:  r.int16(),
		})
	}
	return sides, r.err
}

func marshalBrushSides(sides []*bpb.BSPBrushSide) []byte {
	w := &writer{}
	for _, s := range sides {
		w.uint16(s.GetPlaneNum())
		w.int16(s.GetTexinfo())
	}
	return w.buf.Bytes()
}
