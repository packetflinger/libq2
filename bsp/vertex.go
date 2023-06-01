package bsp

import "github.com/packetflinger/libq2/types"

type Vertex types.Vector3

func (bsp *BSPFile) FetchVertices() []Vertex {
	verts := []Vertex{}
	msg := &bsp.LumpData[VerticesLump].Data
	msg.Index = 0
	quantity := len(bsp.LumpData[VerticesLump].Data.Buffer) / 12
	for i := 0; i < quantity; i++ {
		verts = append(verts, Vertex{
			X: msg.ReadLong(),
			Y: msg.ReadLong(),
			Z: msg.ReadLong(),
		})
	}
	return verts
}
