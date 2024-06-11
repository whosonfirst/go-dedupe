package emitter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/filters"
)

func init() {
	ctx := context.Background()
	RegisterEmitter(ctx, "directory", NewDirectoryEmitter)
}

// DirectoryEmitter implements the `Emitter` interface for crawling records in a directory.
type DirectoryEmitter struct {
	Emitter
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
}

// NewDirectoryEmitter() returns a new `DirectoryEmitter` instance configured by 'uri' in the form of:
//
//	directory://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewDirectoryEmitter(ctx context.Context, uri string) (Emitter, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	idx := &DirectoryEmitter{
		filters: f,
	}

	return idx, nil
}

// WalkURI() walks (crawls) the directory named 'uri' and for each file (not excluded by any filters specified
// when `idx` was created) invokes 'index_cb'.
func (idx *DirectoryEmitter) WalkURI(ctx context.Context, index_cb EmitterCallbackFunc, uri string) error {

	abs_path, err := filepath.Abs(uri)

	if err != nil {
		return fmt.Errorf("Failed to derive absolute path for '%s', %w", uri, err)
	}

	crawl_cb := func(path string, info os.FileInfo) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		if info.IsDir() {
			return nil
		}

		fh, err := ReaderWithPath(ctx, path)

		if err != nil {
			return fmt.Errorf("Failed to create reader for '%s', %w", abs_path, err)
		}

		defer fh.Close()

		if idx.filters != nil {

			ok, err := idx.filters.Apply(ctx, fh)

			if err != nil {
				return fmt.Errorf("Failed to apply filters for '%s', %w", abs_path, err)
			}

			if !ok {
				return nil
			}

			_, err = fh.Seek(0, 0)

			if err != nil {
				return fmt.Errorf("Failed to seek(0, 0) on reader for '%s', %w", abs_path, err)
			}
		}

		err = index_cb(ctx, path, fh)

		if err != nil {
			return fmt.Errorf("Failed to invoke callback for '%s', %w", abs_path, err)
		}

		return nil
	}

	c := crawl.NewCrawler(abs_path)
	err = c.Crawl(crawl_cb)

	if err != nil {
		return fmt.Errorf("Failed to crawl '%s', %w", abs_path, err)
	}

	return nil
}
