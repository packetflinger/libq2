package bsp

import (
	"github.com/packetflinger/libq2/types"
)

const (
	BSPPlaneSize = 20
)

type BSPPlane struct {
	Normal   types.Vector3
	Distance int32
	Type     int32
}

func (bsp *BSPFile) FetchPlanes() []BSPPlane {
	plaincount := bsp.LumpMeta[PlanesLump].length / BSPPlaneSize
	planes := []BSPPlane{}

	msg := &bsp.LumpData[PlanesLump].Data
	msg.Index = 0
	for i := 0; i < int(plaincount); i++ {
		planes = append(planes, BSPPlane{
			Normal: types.Vector3{
				X: msg.ReadLong(),
				Y: msg.ReadLong(),
				Z: msg.ReadLong(),
			},
			Distance: msg.ReadLong(),
			Type:     msg.ReadLong(),
		})
	}
	return planes
}
