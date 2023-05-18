package demo

import (
	"errors"
	"fmt"
	"os"

	m "github.com/packetflinger/libq2/message"
	u "github.com/packetflinger/libq2/util"
)

type DM2File struct {
	Filename      string
	Handle        *os.File
	Position      int
	ParsingFrames bool
	Serverdata    m.ServerData
	Configstrings [m.MaxConfigStrings]m.ConfigString
	Baselines     [m.MaxEntities]m.PackedEntity
	Frames        []m.ServerFrame
	CurrentFrame  *m.ServerFrame
	PrevFrame     *m.ServerFrame
	LumpCallback  func([]byte)
	Callbacks     m.MessageCallbacks
}

func OpenDM2File(f string) (*DM2File, error) {
	if !u.FileExists(f) {
		return nil, errors.New("no such file")
	}

	fp, e := os.Open(f)
	if e != nil {
		return nil, e
	}

	demo := DM2File{
		Filename: f,
		Handle:   fp,
	}

	return &demo, nil
}

func (demo *DM2File) Close() {
	if demo.Handle != nil {
		demo.Handle.Close()
	}
}

func (demo *DM2File) ParseDM2(cb m.MessageCallbacks) error {
	for {
		lump, size, err := nextLump(demo.Handle, int64(demo.Position))
		if err != nil {
			return err
		}
		if size == 0 {
			break
		}
		demo.Position += size

		//err = demo.ParseLump(lump)
		_, err = m.ParseMessageLump(m.NewMessageBuffer(lump), cb)
		if err != nil {
			return err
		}

		// do what we need to with all the messages read
	}
	return nil
}

func nextLump(f *os.File, pos int64) ([]byte, int, error) {
	_, err := f.Seek(pos, 0)
	if err != nil {
		return []byte{}, 0, err
	}

	lumplen := make([]byte, 4)
	_, err = f.Read(lumplen)
	if err != nil {
		return []byte{}, 0, err
	}

	lenbuf := m.MessageBuffer{Buffer: lumplen, Index: 0}
	fmt.Println("lenbuf", lenbuf)
	length := lenbuf.ReadLong()

	// EOF
	if length == -1 {
		return []byte{}, 0, nil
	}

	_, err = f.Seek(pos+4, 0)
	if err != nil {
		return []byte{}, 0, err
	}

	lump := make([]byte, length)
	read, err := f.Read(lump)
	if err != nil {
		return []byte{}, 0, err
	}

	return lump, read + 4, nil
}
