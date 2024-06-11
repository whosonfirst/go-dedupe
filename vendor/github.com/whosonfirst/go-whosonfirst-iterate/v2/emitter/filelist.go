package emitter

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"

	"github.com/whosonfirst/go-whosonfirst-iterate/v2/filters"
)

func init() {
	ctx := context.Background()
	RegisterEmitter(ctx, "filelist", NewFileListEmitter)
}

// FileListEmitter implements the `Emitter` interface for crawling records listed in a "file list" (a plain text newline-delimted list of files).
type FileListEmitter struct {
	Emitter
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
}

// NewFileListEmitter() returns a new `FileListEmitter` instance configured by 'uri' in the form of:
//
//	file://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewFileListEmitter(ctx context.Context, uri string) (Emitter, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	idx := &FileListEmitter{
		filters: f,
	}

	return idx, nil
}

// WalkURI() walks (crawls) the list of files found in 'uri' and for each file (not excluded by any filters specified
// when `idx` was created) invokes 'index_cb'.
func (idx *FileListEmitter) WalkURI(ctx context.Context, index_cb EmitterCallbackFunc, uri string) error {

	abs_path, err := filepath.Abs(uri)

	if err != nil {
		return fmt.Errorf("Failed to derive absolute path for '%s', %w", uri, err)
	}

	fh, err := ReaderWithPath(ctx, abs_path)

	if err != nil {
		return fmt.Errorf("Failed to create reader for '%s', %w", abs_path, err)
	}

	defer fh.Close()

	scanner := bufio.NewScanner(fh)

	for scanner.Scan() {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		path := scanner.Text()

		fh, err := ReaderWithPath(ctx, path)

		if err != nil {
			return fmt.Errorf("Failed to create reader for '%s', %w", path, err)
		}

		if idx.filters != nil {

			ok, err := idx.filters.Apply(ctx, fh)

			if err != nil {
				return fmt.Errorf("Failed to apply filters to '%s', %w", path, err)
			}

			if !ok {
				return nil
			}

			_, err = fh.Seek(0, 0)

			if err != nil {
				return fmt.Errorf("Failed to reset file handle for '%s', %w", path, err)
			}
		}

		err = index_cb(ctx, path, fh)

		if err != nil {
			return fmt.Errorf("Index callback failed for '%s', %w", path, err)
		}
	}

	err = scanner.Err()

	if err != nil {
		return fmt.Errorf("Scanner error reported, %w", err)
	}

	return nil
}
