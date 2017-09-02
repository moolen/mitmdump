package mitmdump

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
)

// LogContext is used to save state between request and response cycles
type LogContext struct {
	// TraceID is a trace id that is used to group multiple
	// requests and responses
	TraceID string
	// UUID is used to identify a single request-response cycle
	UUID string
}

// HTTPLogger can log HTTP requests or responses
type HTTPLogger interface {
	LogReq(req *http.Request, ctx *LogContext) error
	LogRes(res *http.Response, ctx *LogContext) error
}

// FileHTTPLogger implements HTTPLogger and dumps
// the request-response cycle on the file system
// with the following structure:
// {basepath}/{trace-id}/{req-id}/{req|res}
type FileHTTPLogger struct {
	basepath string
	builder  FileStreamBuilder
}

var errEmptyID = errors.New("empty traceID")

// NewLogger creates a FileHTTPLogger that logs
// the request-response cycle uncompressed in plaintext
func NewLogger(basepath string) (*FileHTTPLogger, error) {
	err := os.MkdirAll(basepath, 0755)
	if err != nil {
		return nil, err
	}
	return &FileHTTPLogger{basepath, NewFileStream}, nil
}

// NewZipLogger creates a FileHTTPLogger that logs
// the request-response cycle in gzip compressed format
func NewZipLogger(basepath string, builder FileStreamBuilder) (*FileHTTPLogger, error) {
	err := os.MkdirAll(basepath, 0755)
	if err != nil {
		return nil, err
	}
	return &FileHTTPLogger{basepath, NewGzipFileStream}, nil
}

// LogReq dumps the request to the filesystem
func (logger *FileHTTPLogger) LogReq(req *http.Request, ctx *LogContext) error {
	if ctx.TraceID == "" || ctx.UUID == "" {
		return errEmptyID
	}
	dir := path.Join(logger.basepath, ctx.TraceID, ctx.UUID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	resData, err := httputil.DumpRequest(req, true)
	if err != nil {
		return err
	}
	respath := path.Join(dir, "req")
	stream, err := logger.builder(respath)
	stream.Write(resData)
	stream.Close()
	return nil
}

// LogRes dumps the request to the filesystem
func (logger *FileHTTPLogger) LogRes(res *http.Response, ctx *LogContext) error {
	if ctx.TraceID == "" || ctx.UUID == "" {
		return errEmptyID
	}
	dir := path.Join(logger.basepath, ctx.TraceID, ctx.UUID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	// res might be nil!
	if res == nil {
		return nil
	}
	resData, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}
	respath := path.Join(dir, "res")
	stream, err := logger.builder(respath)
	stream.Write(resData)
	stream.Close()
	return nil
}

// Close does some cleanup if needed
func (logger *FileHTTPLogger) Close() error {
	return nil
}
