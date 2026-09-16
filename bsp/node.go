package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// nodeSize is the on-disk size of a dnode_t: planenum + 2 children (4
// bytes each), a short mins/maxs bounding box (12 bytes), and firstface +
// numfaces (2 bytes each).
const nodeSize = 28

func parseNodes(data []byte) ([]*bpb.BSPNode, error) {
	if len(data)%nodeSize != 0 {
		return nil, fmt.Errorf("nodes lump size %d is not a multiple of %d", len(data), nodeSize)
	}
	count := len(data) / nodeSize
	nodes := make([]*bpb.BSPNode, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		nodes = append(nodes, &bpb.BSPNode{
			PlaneNum:   r.int32Val(),
			FrontChild: r.int32Val(),
			BackChild:  r.int32Val(),
			Bounds:     r.shortBounds(),
			FirstFace:  r.uint16(),
			NumFaces:   r.uint16(),
		})
	}
	return nodes, r.err
}

func marshalNodes(nodes []*bpb.BSPNode) []byte {
	w := &writer{}
	for _, n := range nodes {
		w.int32Val(n.GetPlaneNum())
		w.int32Val(n.GetFrontChild())
		w.int32Val(n.GetBackChild())
		w.shortBounds(n.GetBounds())
		w.uint16(n.GetFirstFace())
		w.uint16(n.GetNumFaces())
	}
	return w.buf.Bytes()
}
