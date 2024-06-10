package writer

import (
	"context"
	"io"
	"log"
	"os"
)

// StdoutWriter is a struct that implements the `Writer` interface for writing documents to STDOUT.
type StdoutWriter struct {
	Writer
}

func init() {

	ctx := context.Background()
	err := RegisterWriter(ctx, "stdout", NewStdoutWriter)

	if err != nil {
		panic(err)
	}
}

// NewStdoutWriter returns a new `CwdWriter` instance for writing documents to STDOUT configured by
// 'uri' in the form of:
//
//	stdout://
//
// Technically 'uri' can also be an empty string.
func NewStdoutWriter(ctx context.Context, uri string) (Writer, error) {

	wr := &StdoutWriter{}
	return wr, nil
}

// Write copies the content of 'fh' to 'path' using an `os.Stdout` writer.
func (wr *StdoutWriter) Write(ctx context.Context, path string, fh io.ReadSeeker) (int64, error) {
	return io.Copy(os.Stdout, fh)
}

// WriterURI returns the value of 'path'
func (wr *StdoutWriter) WriterURI(ctx context.Context, path string) string {
	return path
}

// Flush is a no-op to conform to the `Writer` instance and returns nil.
func (wr *StdoutWriter) Flush(ctx context.Context) error {
	return nil
}

// Close is a no-op to conform to the `Writer` instance and returns nil.
func (wr *StdoutWriter) Close(ctx context.Context) error {
	return nil
}

// SetLogger assigns 'logger' to 'wr'.
func (wr *StdoutWriter) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
