package pak

import (
	"errors"

	"github.com/packetflinger/libq2/message"
	pb "github.com/packetflinger/libq2/proto"
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

func Unmarshal(data []byte) (*pb.PAKArchive, error) {
	header := message.NewMessageBuffer(data[:HeaderLength])
	if header.ReadLong() != Magic {
		return nil, errors.New("not a valid PAK file")
	}
	location := header.ReadLong()
	length := header.ReadLong()
	index := message.NewMessageBuffer(data[location : location+length])
	fileCount := len(index.Buffer) / FileBlockLength
	files := []*pb.PAKFile{}
	for i := 0; i < fileCount; i++ {
		name := index.ReadString()
		index.Index += FileNameLength - len(name) - 1
		dataloc := index.ReadLong()
		datalen := index.ReadLong()
		files = append(files, &pb.PAKFile{
			Name: name,
			Data: data[dataloc : dataloc+datalen],
		})
	}
	pak := &pb.PAKArchive{
		Files: files,
	}
	return pak, nil
}

// generate a binary pak, it should be written to disk after.
func Marshal(archive *pb.PAKArchive) ([]byte, error) {
	buf := message.MessageBuffer{}
	dataLump := message.MessageBuffer{}
	metaLump := message.MessageBuffer{}
	for _, f := range archive.GetFiles() {
		metaLump.WriteString(f.Name)
		for i := len(f.Name); i < FileNameLength-1; i++ {
			metaLump.WriteByte(0) // fill in remaining name space with nulls
		}
		metaLump.WriteLong(int32(dataLump.Index) + HeaderLength)
		metaLump.WriteLong(int32(len(f.Data)))
		dataLump.WriteData(f.Data)
	}
	buf.WriteLong(Magic)
	buf.WriteLong(int32(len(dataLump.Buffer) + HeaderLength))
	buf.WriteLong(int32(len(metaLump.Buffer)))
	buf.Append(dataLump)
	buf.Append(metaLump)
	return buf.Buffer, nil
}

// Add a new file to the contents of the PAK. You'll need to have already
// os.ReadFile()'d to get the data.
func AddFiles(archive *pb.PAKArchive, name string, data []byte) {
	newfile := &pb.PAKFile{
		Name: name,
		Data: data,
	}
	archive.Files = append(archive.Files, newfile)
}

// Delete a file contained in a PAK archive
func RemoveFile(archive *pb.PAKArchive, name string) error {
	files := []*pb.PAKFile{}
	for _, f := range archive.GetFiles() {
		if f.GetName() == name {
			continue
		}
		files = append(files, f)
	}
	if len(files) == len(archive.GetFiles()) {
		return errors.New("file not in PAK archive")
	}
	archive.Files = files
	return nil
}

// Obtain a pointer to a file contained in a PAK archive.
func ExtractFile(archive *pb.PAKArchive, name string) (*pb.PAKFile, error) {
	for _, f := range archive.GetFiles() {
		if f.Name == name {
			return f, nil
		}
	}
	return nil, errors.New("file not found in PAK: " + name)
}
