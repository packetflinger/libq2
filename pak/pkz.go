package pak

import (
	"archive/zip"
	"fmt"
	"os"
)

type PKZFile struct {
	Filename string
	Handle   *os.File
	Size     int64
}

func NewPKZFile(filename string) (*PKZFile, error) {
	pkz := PKZFile{
		Filename: filename,
	}

	fp, err := os.Create(filename)
	if err != nil {
		return &PKZFile{}, err
	}

	pkz.Handle = fp
	return &pkz, nil
}

func OpenPKZFile(filename string) (*PKZFile, error) {
	pkz := PKZFile{
		Filename: filename,
	}

	fp, err := os.Open(filename)
	if err != nil {
		return &PKZFile{}, err
	}

	fi, err := os.Stat(filename)
	if err != nil {
		return &PKZFile{}, err
	}
	pkz.Size = fi.Size()
	pkz.Handle = fp
	return &pkz, nil
}

func (pkz *PKZFile) Close() {
	if pkz.Handle != nil {
		pkz.Handle.Close()
	}
}

func (pkz *PKZFile) ListFiles() ([]string, error) {
	names := []string{}
	reader, err := zip.NewReader(pkz.Handle, pkz.Size)
	if err != nil {
		return names, err
	}
	for _, f := range reader.File {
		names = append(names, f.Name)
	}
	return names, nil
}

// Add a new file to our PKZ archive
func (pkz *PKZFile) AddFile(filepath string) error {
	srcStat, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", filepath)
	}

	writer := zip.NewWriter(pkz.Handle)
	defer writer.Close()
	fh, err := writer.Create(filepath)
	if err != nil {
		return err
	}

	contents, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	_, err = fh.Write(contents)
	if err != nil {
		return err
	}
	return nil
}
