package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// texInfoSize is the on-disk size of a texinfo_t: two Vector3 axes with a
// float offset each (32 bytes), flags + value (8 bytes), a 32-byte fixed
// texture name, and a next_texinfo index (4 bytes).
const texInfoSize = 76

// texInfoNameSize is the fixed width of the texture name field.
const texInfoNameSize = 32

func parseTexInfo(data []byte) ([]*bpb.BSPTexInfo, error) {
	if len(data)%texInfoSize != 0 {
		return nil, fmt.Errorf("texinfo lump size %d is not a multiple of %d", len(data), texInfoSize)
	}
	count := len(data) / texInfoSize
	infos := make([]*bpb.BSPTexInfo, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		infos = append(infos, &bpb.BSPTexInfo{
			SAxis:       r.vector3(),
			SOffset:     r.float32Val(),
			TAxis:       r.vector3(),
			TOffset:     r.float32Val(),
			Flags:       r.int32Val(),
			Value:       r.int32Val(),
			Texture:     r.fixedString(texInfoNameSize),
			NextTexinfo: r.int32Val(),
		})
	}
	return infos, r.err
}

func marshalTexInfo(infos []*bpb.BSPTexInfo) ([]byte, error) {
	w := &writer{}
	for _, t := range infos {
		w.vector3(t.GetSAxis())
		w.float32Val(t.GetSOffset())
		w.vector3(t.GetTAxis())
		w.float32Val(t.GetTOffset())
		w.int32Val(t.GetFlags())
		w.int32Val(t.GetValue())
		if err := w.fixedString(t.GetTexture(), texInfoNameSize); err != nil {
			return nil, err
		}
		w.int32Val(t.GetNextTexinfo())
	}
	return w.buf.Bytes(), nil
}
