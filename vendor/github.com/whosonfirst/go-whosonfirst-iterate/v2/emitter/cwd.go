package emitter

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/filters"
)

var walk_once sync.Once
var walk_error error

func init() {
	ctx := context.Background()
	RegisterEmitter(ctx, "cwd", NewCwdEmitter)
}

// CwdEmitter implements the `Emitter` interface for crawling records in the current (working) directory.
type CwdEmitter struct {
	Emitter
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	cwd     string
}

// NewCwdEmitter() returns a new `CwdEmitter` instance configured by 'uri' in the form of:
//
//	cwd://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewCwdEmitter(ctx context.Context, uri string) (Emitter, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	cwd, err := os.Getwd()

	if err != nil {
		return nil, fmt.Errorf("Failed to derive current working directory, %w", err)
	}

	idx := &CwdEmitter{
		filters: f,
		cwd:     cwd,
	}

	return idx, nil
}

// WalkURI() walks (crawls) the current directory each file (not excluded by any filters specified
// when `idx` was created) invokes 'index_cb'.
func (idx *CwdEmitter) WalkURI(ctx context.Context, index_cb EmitterCallbackFunc, uri string) error {

	walk_once.Do(func() {
		idx.walkOnceFunc(ctx, index_cb)
	})

	if walk_error != nil {
		return walk_error
	}

	return nil
}

func (idx *CwdEmitter) walkOnceFunc(ctx context.Context, index_cb EmitterCallbackFunc) {

	cwd_fs := os.DirFS(idx.cwd)

	walk_error = fs.WalkDir(cwd_fs, ".", func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		r, err := cwd_fs.Open(path)

		if err != nil {
			return fmt.Errorf("Failed to open %s for reading, %w", path, err)
		}

		rsc, err := ioutil.NewReadSeekCloser(r)

		if err != nil {
			return fmt.Errorf("Failed to create ReadSeekCloser for %s, %w", path, err)
		}

		defer rsc.Close()

		if idx.filters != nil {

			ok, err := idx.filters.Apply(ctx, rsc)

			if err != nil {
				return fmt.Errorf("Failed to apply filters for '%s', %w", path, err)
			}

			if !ok {
				return nil
			}

			_, err = rsc.Seek(0, 0)

			if err != nil {
				return fmt.Errorf("Failed to seek(0, 0) on reader for '%s', %w", path, err)
			}
		}

		err = index_cb(ctx, path, rsc)

		if err != nil {
			return fmt.Errorf("Failed to invoke callback for '%s', %w", path, err)
		}

		return nil
	})
}
