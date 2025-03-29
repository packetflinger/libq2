package bsp

type Visibility struct {
	PVS int // visible
	PHS int // audible
}

func (bsp *BSPFile) FetchVisibility() []Visibility {
	vis := []Visibility{}
	visdata := bsp.LumpData[VisibilityLump]
	visCount := len(visdata.Data.Data) / 8

	visdata.Data.Index = 0

	for i := 0; i < visCount; i++ {
		vis = append(vis, Visibility{
			PVS: visdata.Data.ReadLong(),
			PHS: visdata.Data.ReadLong(),
		})
	}

	return vis
}
