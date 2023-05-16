package bsp

import (
	m "github.com/packetflinger/libq2/message"
)

const (
	Magic       = (('P' << 24) + ('S' << 16) + ('B' << 8) + 'I')
	HeaderLen   = 160 // magic + version + lump metadata
	EntityLump  = 0
	TextureLump = 5  // the location in the header
	TextureLen  = 76 // 40 bytes of origins and angles + 36 for textname
)

type BSPFile struct {
}

// Make sure the first 4 bytes match the magic number
func (bsp BSPFile) Validate(header []byte) bool {
	msg := m.MessageBuffer{
		Buffer: header,
	}
	return msg.ReadLong() == Magic
}
