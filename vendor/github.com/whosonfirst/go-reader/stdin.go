package reader

import (
	"context"
	"github.com/whosonfirst/go-ioutil"
	"io"
	"os"
)

// Constant string value representing STDIN.
const STDIN string = "-"

// StdinReader is a struct that implements the `Reader` interface for reading documents from STDIN.
type StdinReader struct {
	Reader
}

func init() {

	ctx := context.Background()
	err := RegisterReader(ctx, "stdin", NewStdinReader)

	if err != nil {
		panic(err)
	}
}

// NewStdinReader returns a new `FileReader` instance for reading documents from STDIN,
// configured by 'uri' in the form of:
//
//	stdin://
//
// Technically 'uri' can also be an empty string.
func NewStdinReader(ctx context.Context, uri string) (Reader, error) {

	r := &StdinReader{}
	return r, nil
}

// Read will open a `io.ReadSeekCloser` instance wrapping `os.Stdin`.
func (r *StdinReader) Read(ctx context.Context, uri string) (io.ReadSeekCloser, error) {
	return ioutil.NewReadSeekCloser(os.Stdin)
}

// ReaderURI will return the value of the `STDIN` constant.
func (r *StdinReader) ReaderURI(ctx context.Context, uri string) string {
	return STDIN
}
