package bsp

const (
	BSPPlaneSize = 20
)

type BSPPlane struct {
	Data []byte
}

func (bsp *BSPFile) FetchPlanes() []BSPPlane {
	plaincount := bsp.LumpMeta[PlanesLump].length / BSPPlaneSize
	planes := []BSPPlane{}

	for i := 0; i < int(plaincount); i++ {
		planes = append(planes, BSPPlane{
			Data: bsp.LumpData[PlanesLump].Data.ReadData(20),
		})
	}
	return planes
}
