package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// visOffsetSize is the on-disk size of one dvis_t bitofs[] entry: a pvs and
// a phs int32 offset.
const visOffsetSize = 8

// parseVisibility decodes the dvis_t structure: a cluster count, followed
// by that many (pvs, phs) offset pairs, followed by the raw RLE-encoded
// bit vectors those offsets point into. The offsets are stored exactly as
// read (they're absolute offsets from the start of this lump), so they
// remain valid against BitVectors reconstructed by marshalVisibility.
func parseVisibility(data []byte) (*bpb.BSPVisibility, error) {
	if len(data) == 0 {
		return &bpb.BSPVisibility{}, nil
	}

	r := newReader(data)
	clusterCount := r.int32Val()
	if r.err != nil {
		return nil, r.err
	}

	offsets := make([]*bpb.BSPVisOffset, 0, clusterCount)
	for i := int32(0); i < clusterCount; i++ {
		offsets = append(offsets, &bpb.BSPVisOffset{
			Pvs: r.int32Val(),
			Phs: r.int32Val(),
		})
	}
	if r.err != nil {
		return nil, fmt.Errorf("reading vis offset table: %w", r.err)
	}

	bitVectors := append([]byte{}, data[r.pos:]...)

	return &bpb.BSPVisibility{
		ClusterCount: clusterCount,
		Offsets:      offsets,
		BitVectors:   bitVectors,
	}, nil
}

func marshalVisibility(vis *bpb.BSPVisibility) []byte {
	if vis == nil || vis.GetClusterCount() == 0 {
		return nil
	}
	w := &writer{}
	w.int32Val(vis.GetClusterCount())
	for _, off := range vis.GetOffsets() {
		w.int32Val(off.GetPvs())
		w.int32Val(off.GetPhs())
	}
	w.bytes(vis.GetBitVectors())
	return w.buf.Bytes()
}
