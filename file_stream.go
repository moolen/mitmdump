package mitmdump

import (
	"compress/gzip"
	"errors"
	"io"
	"os"
)

// FileStream is a WriteCloser that writes to a file
type FileStream struct {
	path string
	f    io.WriteCloser
}

// The FileStreamBuilder is used as a factory
// function in order to create a FileStream
type FileStreamBuilder func(string) (*FileStream, error)

// NewFileStream returns a FileStream that writes plain text
// to the specified path
func NewFileStream(path string) (*FileStream, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &FileStream{path, file}, nil
}

// NewGzipFileStream returns a new FileStream that writes
// compressed data to the specified path
func NewGzipFileStream(path string) (*FileStream, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	writeCloser := gzip.NewWriter(file)
	return &FileStream{path, writeCloser}, nil
}

// implements Writer
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
	if fs.f == nil {
		return errors.New("FileStream was never written into")
	}
	return fs.f.Close()
}
