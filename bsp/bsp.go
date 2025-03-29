package bsp

import (
	"errors"
	"os"
	"strings"

	m "github.com/packetflinger/libq2/message"
	u "github.com/packetflinger/libq2/util"
)

const (
	Magic          = (('P' << 24) + ('S' << 16) + ('B' << 8) + 'I')
	HeaderLen      = 160 // magic + version + lump metadata
	EntityLump     = 0
	PlanesLump     = 1 // 20 bytes each
	VerticesLump   = 2 // 12 bytes each (3 floats)
	VisibilityLump = 3 // 4 bytes PVS, 4 bytes PHS
	NodesLump      = 4 // 28 bytes each
	TextureLump    = 5 // 76 bytes each
	FacesLump      = 6 // 20 bytes each
	LightMapLump   = 7
	LeavesLump     = 8
	LeafFaceTable  = 9
	LeafBrushTable = 10
	EdgesLump      = 11
	FaceEdgeTable  = 12
	ModelsLump     = 13
	BrushesLump    = 14
	BrushSidesLump = 15
	POPLump        = 16
	AreasLump      = 17
	AreaPortals    = 18
	TextureLen     = 76 // 40 bytes of origins and angles + 36 for textname
)

// represents a binary space partitioning map file
type BSPFile struct {
	Name       string // just the filename minus extension (ex: "q2dm1")
	Filename   string // including any path prefix and .bsp extension
	Handle     *os.File
	Header     m.Buffer
	LumpMeta   [19]BSPLumpMeta
	LumpData   [19]BSPLumpData
	Ents       []BSPEntity
	Planes     []BSPPlane
	Vertices   []Vertex
	Visibility []Visibility
}

// Collections of data are organized into "lumps" within the file
type BSPLumpMeta struct {
	location int
	length   int
}

type BSPLumpData struct {
	Data m.Buffer
}

func OpenBSPFile(f string) (*BSPFile, error) {
	if !u.FileExists(f) {
		return nil, errors.New("no such file")
	}

	fp, e := os.Open(f)
	if e != nil {
		return nil, e
	}

	header := make([]byte, HeaderLen)
	_, e = fp.Read(header)
	if e != nil {
		return nil, e
	}
	tokens := strings.Split(f, string(os.PathSeparator))
	ftokens := strings.Split(tokens[len(tokens)-1], ".")

	bsp := BSPFile{
		Name:     ftokens[0],
		Filename: f,
		Handle:   fp,
		Header:   m.NewBuffer(header),
	}

	bsp.ParseLumpMetadata()
	e = bsp.ParseLumpData()
	if e != nil {
		return nil, e
	}

	bsp.Ents = bsp.FetchEntities()
	bsp.Planes = bsp.FetchPlanes()
	bsp.Vertices = bsp.FetchVertices()
	bsp.Visibility = bsp.FetchVisibility()

	return &bsp, nil
}

func (b *BSPFile) Close() {
	if b.Handle != nil {
		b.Handle.Close()
	}
}

// Make sure the first 4 bytes match the magic number
func (bsp *BSPFile) Validate() bool {
	bsp.Header.Index = 0
	return bsp.Header.ReadLong() == Magic
}

// find all the locations/sizes of the various lumps. There are
// always 18 of them.
func (bsp *BSPFile) ParseLumpMetadata() {
	bsp.Header.Index = 8 // 4 for magic + 4 for meta location
	for i := 0; i < 18; i++ {
		bsp.LumpMeta[i].location = bsp.Header.ReadLong()
		bsp.LumpMeta[i].length = bsp.Header.ReadLong()
	}
}

func (bsp *BSPFile) ParseLumpData() error {
	for i := 0; i < 18; i++ {
		_, err := bsp.Handle.Seek(int64(bsp.LumpMeta[i].location), 0)
		if err != nil {
			return err
		}

		data := make([]byte, int(bsp.LumpMeta[i].length))
		read, err := bsp.Handle.Read(data)
		if err != nil {
			return err
		}

		if read != int(bsp.LumpMeta[i].length) {
			return errors.New("reading texture lump: hit EOF")
		}

		bsp.LumpData[i] = BSPLumpData{
			Data: m.Buffer{
				Data: data,
			},
		}
	}
	return nil
}
