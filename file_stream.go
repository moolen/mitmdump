package mitmdump

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
)

// FileStream is a WriteCloser that writes to a file
type FileStream struct {
	path string
	f    io.WriteCloser
}

// NewFileStream returns a new FileStream in "as-is" mode
func NewFileStream(path string) (*FileStream, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &FileStream{path, file}, nil
}

// NewGzipFileStream returns a new FileStream in "gzip" mode
func NewGzipFileStream(path string) (*FileStream, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	writeCloser := gzip.NewWriter(file)
	return &FileStream{path, writeCloser}, nil
}

// Write implements Writer
func (fs *FileStream) Write(b []byte) (nr int, err error) {
	if fs.f == nil {
		file, err := os.Create(fs.path)
		if err != nil {
			return 0, err
		}
		fs.f = gzip.NewWriter(file)
	}
	return fs.f.Write(b)
}

// Close implements Closer
func (fs *FileStream) Close() error {
	fmt.Println("Close", fs.path)
	if fs.f == nil {
		return errors.New("FileStream was never written into")
	}
	return fs.f.Close()
}
