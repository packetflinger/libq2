package bsp

import (
	"errors"
	"os"
	"strings"

	m "github.com/packetflinger/libq2/message"
	u "github.com/packetflinger/libq2/util"
)

const (
	Magic       = (('P' << 24) + ('S' << 16) + ('B' << 8) + 'I')
	HeaderLen   = 160 // magic + version + lump metadata
	EntityLump  = 0
	TextureLump = 5  // the location in the header
	TextureLen  = 76 // 40 bytes of origins and angles + 36 for textname
)

type BSPFile struct {
	Name     string // just the filename minus extension (ex: "q2dm1")
	Filename string // including any path prefix and .bsp extension
	Handle   *os.File
	Header   m.MessageBuffer
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
		Header:   m.NewMessageBuffer(header),
	}

	return &bsp, nil
}

func (b *BSPFile) Close() {
	if b.Handle != nil {
		b.Handle.Close()
	}
}

// Make sure the first 4 bytes match the magic number
func (bsp *BSPFile) Validate() bool {
	return bsp.Header.ReadLong() == Magic
}
