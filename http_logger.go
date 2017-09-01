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
	TraceID string
	UUID    string
}

// HTTPLogger implements RequestLogger and ResponseLogger
type HTTPLogger struct {
	basepath string
}

var errEmptyID = errors.New("empty traceID")

// NewLogger creates a HTTPLogger
func NewLogger(basepath string) (*HTTPLogger, error) {
	err := os.MkdirAll(basepath, 0755)
	if err != nil {
		return nil, err
	}
	return &HTTPLogger{basepath}, nil
}

// LogReq ...
func (logger *HTTPLogger) LogReq(req *http.Request, ctx *LogContext) error {
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
	stream, err := NewFileStream(respath)
	stream.Write(resData)
	stream.Close()
	return nil
}

// LogRes ..
func (logger *HTTPLogger) LogRes(res *http.Response, ctx *LogContext) error {
	if ctx.TraceID == "" || ctx.UUID == "" {
		return errEmptyID
	}
	dir := path.Join(logger.basepath, ctx.TraceID, ctx.UUID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	resData, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}
	respath := path.Join(dir, "res")
	stream, err := NewFileStream(respath)
	stream.Write(resData)
	stream.Close()
	return nil
}

// Close gracefully shuts down the logger
func (logger *HTTPLogger) Close() error {
	return nil
}
