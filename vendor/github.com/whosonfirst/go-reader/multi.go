package reader

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"io"
	_ "log"
	"sync"
)

// MultiReader is a struct that implements the `Reader` interface for reading documents from one or more `Reader` instances.
type MultiReader struct {
	Reader
	readers []Reader
	lookup  map[string]int
	mu      *sync.RWMutex
}

// NewMultiReaderFromURIs returns a new `Reader` instance for reading documents from one or more `Reader` instances.
// 'uris' is assumed to be a list of URIs each of which will be used to invoke the `NewReader` method.
func NewMultiReaderFromURIs(ctx context.Context, uris ...string) (Reader, error) {

	readers := make([]Reader, 0)

	for _, uri := range uris {

		r, err := NewReader(ctx, uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create reader for %s, %w", uri, err)
		}

		readers = append(readers, r)
	}

	return NewMultiReader(ctx, readers...)
}

// NewMultiReaderFromURIs returns a new `Reader` instance for reading documents from one or more `Reader` instances.
func NewMultiReader(ctx context.Context, readers ...Reader) (Reader, error) {

	lookup := make(map[string]int)

	mu := new(sync.RWMutex)

	mr := MultiReader{
		readers: readers,
		lookup:  lookup,
		mu:      mu,
	}

	return &mr, nil
}

// Read will open an `io.ReadSeekCloser` for a file matching 'path'. In the case of multiple underlying
// `Reader` instances the first instance to successfully load 'path' will be returned.
func (mr *MultiReader) Read(ctx context.Context, path string) (io.ReadSeekCloser, error) {

	missing := errors.New("Unable to read URI")

	mr.mu.RLock()

	idx, ok := mr.lookup[path]

	mr.mu.RUnlock()

	if ok {

		// log.Printf("READ MULTIREADER LOOKUP INDEX FOR %s AS %d\n", path, idx)

		if idx == -1 {
			return nil, missing
		}

		r := mr.readers[idx]
		return r.Read(ctx, path)
	}

	var fh io.ReadSeekCloser
	idx = -1

	var errors error

	for i, r := range mr.readers {

		rsp, err := r.Read(ctx, path)

		if err != nil {
			errors = multierror.Append(fmt.Errorf("Failed to read %s with %T, %w", path, r, err))
		} else {

			fh = rsp
			idx = i

			break
		}
	}

	// log.Printf("SET MULTIREADER LOOKUP INDEX FOR %s AS %d\n", path, idx)

	mr.mu.Lock()
	mr.lookup[path] = idx
	mr.mu.Unlock()

	if fh == nil {
		return nil, errors
	}

	return fh, nil
}

// ReaderURI returns the absolute URL for 'path'. In the case of multiple underlying
// `Reader` instances the first instance to successfully load 'path' will be returned.
func (mr *MultiReader) ReaderURI(ctx context.Context, path string) string {

	mr.mu.RLock()

	idx, ok := mr.lookup[path]

	mr.mu.RUnlock()

	if ok {
		return mr.readers[idx].ReaderURI(ctx, path)
	}

	r, err := mr.Read(ctx, path)

	if err != nil {
		return fmt.Sprintf("x-urn:go-reader:multi#%s", path)
	}

	defer r.Close()

	return mr.ReaderURI(ctx, path)
}
