package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// faceSize is the on-disk size of a dface_t: planenum (2 bytes), side (2
// bytes), firstedge (4 bytes), numedges (2 bytes), texinfo (2 bytes), 4
// lightmap style bytes, and a lightofs int32 (4 bytes).
const faceSize = 20

// faceLightStyles is how many lightmap style bytes each face has.
const faceLightStyles = 4

func parseFaces(data []byte) ([]*bpb.BSPFace, error) {
	if len(data)%faceSize != 0 {
		return nil, fmt.Errorf("faces lump size %d is not a multiple of %d", len(data), faceSize)
	}
	count := len(data) / faceSize
	faces := make([]*bpb.BSPFace, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		f := &bpb.BSPFace{
			PlaneNum:  r.uint16(),
			Side:      r.int16(),
			FirstEdge: r.int32Val(),
			NumEdges:  r.int16(),
			Texinfo:   r.int16(),
		}
		f.LightStyles = make([]uint32, faceLightStyles)
		for j := range f.LightStyles {
			f.LightStyles[j] = r.byteVal()
		}
		f.LightOffset = r.int32Val()
		faces = append(faces, f)
	}
	return faces, r.err
}

func marshalFaces(faces []*bpb.BSPFace) ([]byte, error) {
	w := &writer{}
	for _, f := range faces {
		w.uint16(f.GetPlaneNum())
		w.int16(f.GetSide())
		w.int32Val(f.GetFirstEdge())
		w.int16(f.GetNumEdges())
		w.int16(f.GetTexinfo())
		styles := f.GetLightStyles()
		if len(styles) != faceLightStyles {
			return nil, fmt.Errorf("face has %d light styles, want exactly %d", len(styles), faceLightStyles)
		}
		for _, s := range styles {
			w.byteVal(s)
		}
		w.int32Val(f.GetLightOffset())
	}
	return w.buf.Bytes(), nil
}
