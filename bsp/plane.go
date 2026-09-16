package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// planeSize is the on-disk size of a dplane_t: a Vector3 normal, a float
// distance, and an int32 type.
const planeSize = 20

func parsePlanes(data []byte) ([]*bpb.BSPPlane, error) {
	if len(data)%planeSize != 0 {
		return nil, fmt.Errorf("planes lump size %d is not a multiple of %d", len(data), planeSize)
	}
	count := len(data) / planeSize
	planes := make([]*bpb.BSPPlane, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		planes = append(planes, &bpb.BSPPlane{
			Normal:   r.vector3(),
			Distance: r.float32Val(),
			Type:     r.int32Val(),
		})
	}
	return planes, r.err
}

func marshalPlanes(planes []*bpb.BSPPlane) ([]byte, error) {
	w := &writer{}
	for _, p := range planes {
		w.vector3(p.GetNormal())
		w.float32Val(p.GetDistance())
		w.int32Val(p.GetType())
	}
	return w.buf.Bytes(), nil
}
