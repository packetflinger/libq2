package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// modelSize is the on-disk size of a dmodel_t: a float mins/maxs bounding
// box (24 bytes), an origin Vector3 (12 bytes), and headnode + firstface +
// numfaces (4 bytes each).
const modelSize = 48

func parseModels(data []byte) ([]*bpb.BSPModel, error) {
	if len(data)%modelSize != 0 {
		return nil, fmt.Errorf("models lump size %d is not a multiple of %d", len(data), modelSize)
	}
	count := len(data) / modelSize
	models := make([]*bpb.BSPModel, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		models = append(models, &bpb.BSPModel{
			Bounds:    r.floatBounds(),
			Origin:    r.vector3(),
			HeadNode:  r.int32Val(),
			FirstFace: r.int32Val(),
			NumFaces:  r.int32Val(),
		})
	}
	return models, r.err
}

func marshalModels(models []*bpb.BSPModel) []byte {
	w := &writer{}
	for _, m := range models {
		w.floatBounds(m.GetBounds())
		w.vector3(m.GetOrigin())
		w.int32Val(m.GetHeadNode())
		w.int32Val(m.GetFirstFace())
		w.int32Val(m.GetNumFaces())
	}
	return w.buf.Bytes()
}
