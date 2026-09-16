package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// vertexSize is the on-disk size of a dvertex_t: a single Vector3 point.
const vertexSize = 12

func parseVertices(data []byte) ([]*bpb.BSPVertex, error) {
	if len(data)%vertexSize != 0 {
		return nil, fmt.Errorf("vertices lump size %d is not a multiple of %d", len(data), vertexSize)
	}
	count := len(data) / vertexSize
	verts := make([]*bpb.BSPVertex, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		verts = append(verts, &bpb.BSPVertex{Point: r.vector3()})
	}
	return verts, r.err
}

func marshalVertices(verts []*bpb.BSPVertex) []byte {
	w := &writer{}
	for _, v := range verts {
		w.vector3(v.GetPoint())
	}
	return w.buf.Bytes()
}
