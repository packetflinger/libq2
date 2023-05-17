package pak

import (
	"errors"
	"fmt"
	"os"

	m "github.com/packetflinger/libq2/message"
	u "github.com/packetflinger/libq2/util"
)

const (
	Magic           = (('K' << 24) + ('C' << 16) + ('A' << 8) + 'P')
	HeaderLength    = 12
	FileBlockLength = 64 // name + offset + length
	FileNameLength  = 56
	FileOffset      = 56
	FileLength      = 60
	Separator       = "/" // always use linux-style, even on windows
)

// A .pak archive
type PakFile struct {
	Filename string
	Handle   *os.File
	Size     uint64
	Header   m.MessageBuffer
	Index    PakFileIndex
	Files    []PackedFile
}

type PakFileIndex struct {
	Location uint32
	Length   uint32
}

// a file contained inside a pak file
type PackedFile struct {
	Name   string
	Offset int
	Length int
	Data   []byte
}

func OpenPakFile(f string) (*PakFile, error) {
	if !u.FileExists(f) {
		return nil, errors.New("no such file")
	}

	fp, e := os.Open(f)
	if e != nil {
		return nil, e
	}

	header := make([]byte, HeaderLength)
	_, e = fp.Read(header)
	if e != nil {
		return nil, e
	}

	pak := PakFile{
		Filename: f,
		Handle:   fp,
		Header:   m.NewMessageBuffer(header),
	}

	if !pak.Validate() {
		return nil, errors.New("invalid pak file")
	}

	idx := PakFileIndex{
		Location: uint32(pak.Header.ReadLong()),
		Length:   uint32(pak.Header.ReadLong()),
	}
	pak.Index = idx

	e = pak.ParseFileIndex()
	if e != nil {
		return nil, e
	}

	e = pak.ParseFileData()
	if e != nil {
		return nil, e
	}

	return &pak, nil
}

func (pak *PakFile) Close() {
	if pak.Handle != nil {
		pak.Handle.Close()
	}
}

// Make sure the first 4 bytes match the magic number
func (pak *PakFile) Validate() bool {
	pak.Header.Index = 0
	return pak.Header.ReadLong() == Magic
}

// the end of a pak file contains an index to all the files contained
func (pak *PakFile) ParseFileIndex() error {
	_, e := pak.Handle.Seek(int64(pak.Index.Location), 0)
	if e != nil {
		return e
	}

	indexBlock := make([]byte, pak.Index.Length)
	_, e = pak.Handle.Read(indexBlock)
	if e != nil {
		return e
	}

	filecount := pak.Index.Length / FileBlockLength
	block := make([]byte, FileBlockLength)

	for i := 0; i < int(filecount); i++ {
		pFile := PackedFile{}
		_, e = pak.Handle.Seek(int64(int(pak.Index.Location)+(i*FileBlockLength)), 0)
		if e != nil {
			return e
		}
		_, e = pak.Handle.Read(block)
		if e != nil {
			return e
		}
		msg := m.MessageBuffer{
			Buffer: block,
		}
		pFile.Name = msg.ReadString()
		msg.Index = FileOffset
		pFile.Offset = int(msg.ReadLong())
		pFile.Length = int(msg.ReadLong())
		pak.Files = append(pak.Files, pFile)
	}
	return nil
}

func (pak *PakFile) ParseFileData() error {
	for i := range pak.Files {
		_, e := pak.Handle.Seek(int64(pak.Files[i].Offset), 0)
		if e != nil {
			return e
		}

		blob := make([]byte, pak.Files[i].Length)
		_, e = pak.Handle.Read(blob)
		if e != nil {
			return e
		}
		pak.Files[i].Data = blob
	}
	return nil
}

func (pak *PakFile) AddFile(f string) error {
	for _, pf := range pak.Files {
		if pf.Name == f {
			return fmt.Errorf("%s already exists in pak", pf.Name)
		}
	}

	data, e := os.ReadFile(f)
	if e != nil {
		return e
	}

	pak.Files = append(pak.Files, PackedFile{
		Name:   f,
		Length: len(data),
		Data:   data,
		// we don't care about setting an offset
	})
	return nil
}

func (pak *PakFile) RemoveFile(f string) error {
	target := -1
	for i, pf := range pak.Files {
		if pf.Name == f {
			target = i
		}
	}

	if target == -1 {
		return errors.New("file not found in pak")
	}

	// remove the file from the slice
	filelist := pak.Files[:target-1]
	filelist = append(filelist, pak.Files[target:]...)
	pak.Files = filelist
	return nil
}
