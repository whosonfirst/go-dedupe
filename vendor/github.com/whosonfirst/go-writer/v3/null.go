package writer

import (
	"context"
	"io"
	"log"
)

// NullWriter is a struct that implements the `Writer` interface for writing documents to nowhere.
type NullWriter struct {
	Writer
}

func init() {

	ctx := context.Background()
	err := RegisterWriter(ctx, "null", NewNullWriter)

	if err != nil {
		panic(err)
	}
}

// NewNullWriter returns a new `CwdWriter` instance for writing documents to nowhere configured by
// 'uri' in the form of:
//
//	null://
//
// Technically 'uri' can also be an empty string.
func NewNullWriter(ctx context.Context, uri string) (Writer, error) {

	wr := &NullWriter{}
	return wr, nil
}

// Write copies the content of 'fh' to 'path' using an `io.Discard` writer.
func (wr *NullWriter) Write(ctx context.Context, path string, fh io.ReadSeeker) (int64, error) {
	return io.Copy(io.Discard, fh)
}

// WriterURI returns the value of 'path'
func (wr *NullWriter) WriterURI(ctx context.Context, path string) string {
	return path
}

// Flush is a no-op to conform to the `Writer` instance and returns nil.
func (wr *NullWriter) Flush(ctx context.Context) error {
	return nil
}

// Close is a no-op to conform to the `Writer` instance and returns nil.
func (wr *NullWriter) Close(ctx context.Context) error {
	return nil
}

// SetLogger is a no-op to conform to the `Writer` instance and returns nil.
func (wr *NullWriter) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
