// A .pak file is a container for a collection of files similar to a .zip but
// with no compression. Used for implementing a very simple virual filesystem
// in Quake 2. Multiple PAK files are usually defined and their contents are
// combined into a single virtual filesystem.
//
// All Quake 2 clients support PAK files, as it was the original archive format
// included when the game was created. The q2pro client supports .pkz files
// which are basically just .zip files with a different extension. This is the
// much preferred archive format as it's natively supported in nearly all OSes.
//
// This implementation does not allow for duplicate files in the same archive.
// If you try to add a file of the same name or data as an existing file in the
// pak an error will surface.
//
// Structure -
// PAK files consist of a header, a blob of all the file contents concatinated
// together, and an index of the file paths and locations where the data for
// those files can be found within the blob.
//
//	Entire file:
//	  [ header ][ all file contents blob ][ index ]
//	Header:
//	  [ magic 4-bytes ][ index location 4-bytes ][ index length 4-bytes ]
//	Index:
//	  [ filename 56-bytes ][ data location 4-bytes ][ data len 4 bytes ]
//
// Limitations -
// File paths can only be upto 56 characters long. An index entry is 64 bytes
// with the data location being 4 bytes and the data length also being 4 bytes.
// The remaining 56 byte can be used for the file's path/name. Any unused space
// in the name is padded with nulls.
//
// There is a hard-limit 4GB for a PAK file. This limition applies for both the
// overall size of the .pak file and any individual file contained inside the
// container. This is due to the internal addressing using 32-bit integers for
// data lengths and locations. The header contains the offset for where the
// index is located at the end of the file; that offset at can at most be
// ~4,294,000,000 bytes +/- depending on the number of files in the index.
package pak

import (
	"crypto/sha256"
	"encoding/hex"
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
	MinLength       = 77  // header + index with 1 file, single byte of data
)

var (
	ErrorDuplicateData = errors.New("duplicate file data found in pak")
	ErrorDuplicateName = errors.New("duplicate file name in pak")
	ErrorFileNotFound  = errors.New("file not found in pak")
	ErrorInvalidPak    = errors.New("invalid pak archive")
	ErrorNilPak        = errors.New("nil pak archive")
	ErrorShortPak      = errors.New("invalid pak file, not enough data")
)

// Load the binary .pak file data into a structured text proto. This structure
// can be easily modified (changing/adding/removing files) and then changed
// back into a binary .pak file.
func Unmarshal(data []byte) (*pb.PAKArchive, error) {
	if len(data) < MinLength {
		return nil, ErrorShortPak
	}
	header := message.NewBuffer(data[:HeaderLength])
	if header.ReadLong() != Magic {
		return nil, ErrorInvalidPak
	}
	location := header.ReadLong()
	length := header.ReadLong()
	index := message.NewBuffer(data[location : location+length])
	fileCount := len(index.Data) / FileBlockLength
	files := []*pb.PAKFile{}
	for range fileCount {
		name := index.ReadString()
		index.Index += FileNameLength - len(name) - 1
		dataloc := index.ReadLong()
		datalen := index.ReadLong()
		content := data[dataloc : dataloc+datalen]
		files = append(files, &pb.PAKFile{
			Name:     name,
			Data:     content,
			Location: int32(dataloc),
			Length:   int32(datalen),
			Hash:     dataHash(content),
		})
	}
	pak := &pb.PAKArchive{
		Files: files,
	}
	return pak, nil
}

// Generate a binary .pak from the text proto. The data can then be written to
// to disk for a fully materialized .pak file.
func Marshal(archive *pb.PAKArchive) ([]byte, error) {
	if archive == nil {
		return []byte{}, ErrorNilPak
	}
	var buf, dataLump, metaLump message.Buffer
	names := make(map[string]bool)
	for _, f := range archive.GetFiles() {
		metaLump.WriteString(f.Name)
		for i := len(f.Name); i < FileNameLength-1; i++ {
			metaLump.WriteByte(0) // fill in remaining name space with nulls
		}
		metaLump.WriteLong(dataLump.Index + HeaderLength)
		metaLump.WriteLong(len(f.Data))
		dataLump.WriteData(f.Data)
		if _, dup := names[f.Name]; dup {
			return []byte{}, ErrorDuplicateName
		}
		names[f.Name] = true
	}
	buf.WriteLong(Magic)
	buf.WriteLong(len(dataLump.Data) + HeaderLength)
	buf.WriteLong(len(metaLump.Data))
	buf.Append(dataLump)
	buf.Append(metaLump)
	return buf.Data, nil
}

// Add a new file virtual file to the contents of the PAK.
func AddFile(archive *pb.PAKArchive, name string, data []byte) error {
	if archive == nil {
		return ErrorNilPak
	}
	hash := dataHash(data)
	for _, f := range archive.GetFiles() {
		if f.GetName() == name {
			return ErrorDuplicateName
		}
		if f.GetHash() == hash {
			return ErrorDuplicateData
		}
	}
	newfile := &pb.PAKFile{
		Name: name,
		Data: data,
	}
	archive.Files = append(archive.Files, newfile)
	return nil
}

// Delete a file contained in a PAK archive
func RemoveFile(archive *pb.PAKArchive, name string) error {
	if archive == nil {
		return ErrorNilPak
	}
	files := []*pb.PAKFile{}
	for _, f := range archive.GetFiles() {
		if f.GetName() == name {
			continue
		}
		files = append(files, f)
	}
	if len(files) == len(archive.GetFiles()) {
		return ErrorFileNotFound
	}
	archive.Files = files
	return nil
}

// Obtain a pointer to a file contained in a PAK archive.
func ExtractFile(archive *pb.PAKArchive, name string) (*pb.PAKFile, error) {
	if archive == nil {
		return nil, ErrorNilPak
	}
	for _, f := range archive.GetFiles() {
		if f.Name == name {
			return f, nil
		}
	}
	return nil, ErrorFileNotFound
}

// Get an SHA256 hash of the input parameter data
func dataHash(data []byte) string {
	hashbytes := sha256.Sum256(data)
	return hex.EncodeToString(hashbytes[:])
}
