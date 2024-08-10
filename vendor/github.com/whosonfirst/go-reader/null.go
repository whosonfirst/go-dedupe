package reader

import (
	"bytes"
	"context"
	"github.com/whosonfirst/go-ioutil"
	"io"
)

// NullReader is a struct that implements the `Reader` interface for reading documents from nowhere.
type NullReader struct {
	Reader
}

func init() {

	ctx := context.Background()
	err := RegisterReader(ctx, "null", NewNullReader)

	if err != nil {
		panic(err)
	}
}

// NewNullReader returns a new `FileReader` instance for reading documents from nowhere,
// configured by 'uri' in the form of:
//
//	null://
//
// Technically 'uri' can also be an empty string.
func NewNullReader(ctx context.Context, uri string) (Reader, error) {

	r := &NullReader{}
	return r, nil
}

// Read will open and return an empty `io.ReadSeekCloser` for any value of 'path'.
func (r *NullReader) Read(ctx context.Context, path string) (io.ReadSeekCloser, error) {
	br := bytes.NewReader([]byte(""))
	return ioutil.NewReadSeekCloser(br)
}

// ReaderURI returns the value of 'path'.
func (r *NullReader) ReaderURI(ctx context.Context, path string) string {
	return path
}
