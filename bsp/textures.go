package bsp

type BSPTexture struct {
	File string // max 64 chars
}

// Get a slice of textures
// TODO: parse the location and orientation data too
func (bsp *BSPFile) FetchTextures() []BSPTexture {
	lump := bsp.LumpData[TextureLump].Data
	qty := len(lump.Buffer) / TextureLen
	var textures []BSPTexture
	lump.Index = 0
	for i := 0; i < qty; i++ {
		_ = lump.ReadData(40) // skip texture position and orientation for now
		tdata := lump.ReadData(32)
		_ = lump.ReadLong() // eat up the ending 0xffffffff
		textures = append(textures, BSPTexture{File: string(tdata)})
	}
	return textures
}
