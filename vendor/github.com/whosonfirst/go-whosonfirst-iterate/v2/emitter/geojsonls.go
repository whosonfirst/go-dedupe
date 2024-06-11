package emitter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/filters"
)

func init() {
	ctx := context.Background()
	RegisterEmitter(ctx, "geojsonl", NewGeoJSONLEmitter)
}

// GeojsonLEmitter implements the `Emitter` interface for crawling features in a line-separated GeoJSON record.
type GeojsonLEmitter struct {
	Emitter
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
}

// NewGeojsonLEmitter() returns a new `GeojsonLEmitter` instance configured by 'uri' in the form of:
//
//	geojsonl://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewGeoJSONLEmitter(ctx context.Context, uri string) (Emitter, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	idx := &GeojsonLEmitter{
		filters: f,
	}

	return idx, nil
}

// WalkURI() walks (crawls) each GeoJSON feature found in the file identified by 'uri' and for
// each file (not excluded by any filters specified when `idx` was created) invokes 'index_cb'.
func (idx *GeojsonLEmitter) WalkURI(ctx context.Context, index_cb EmitterCallbackFunc, uri string) error {

	fh, err := ReaderWithPath(ctx, uri)

	if err != nil {
		return fmt.Errorf("Failed to create reader for '%s', %w", uri, err)
	}

	defer fh.Close()

	// see this - we're using ReadLine because it's entirely possible
	// that the raw GeoJSON (LS) will be too long for bufio.Scanner
	// see also - https://golang.org/pkg/bufio/#Reader.ReadLine
	// (20170822/thisisaaronland)

	reader := bufio.NewReader(fh)
	raw := bytes.NewBuffer([]byte(""))

	i := 0

	for {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		path := fmt.Sprintf("%s#%d", uri, i)
		i += 1

		fragment, is_prefix, err := reader.ReadLine()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("Failed to read line at '%s', %w", path, err)
		}

		raw.Write(fragment)

		if is_prefix {
			continue
		}

		br := bytes.NewReader(raw.Bytes())
		fh, err := ioutil.NewReadSeekCloser(br)

		if err != nil {
			return fmt.Errorf("Failed to create new ReadSeekCloser for '%s', %w", path, err)
		}

		defer fh.Close()

		if idx.filters != nil {

			ok, err := idx.filters.Apply(ctx, fh)

			if err != nil {
				return fmt.Errorf("Failed to apply filters for '%s', %w", path, err)
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

		raw.Reset()
	}

	return nil
}
