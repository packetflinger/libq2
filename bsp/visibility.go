package bsp

type Visibility struct {
	PVS int32 // visible
	PHS int32 // audible
}

func (bsp *BSPFile) FetchVisibility() []Visibility {
	vis := []Visibility{}
	visdata := bsp.LumpData[VisibilityLump]
	visCount := len(visdata.Data.Buffer) / 8

	visdata.Data.Index = 0

	for i := 0; i < visCount; i++ {
		vis = append(vis, Visibility{
			PVS: visdata.Data.ReadLong(),
			PHS: visdata.Data.ReadLong(),
		})
	}

	return vis
}
