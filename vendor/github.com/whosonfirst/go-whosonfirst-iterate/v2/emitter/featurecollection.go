package emitter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/filters"
)

func init() {
	ctx := context.Background()
	RegisterEmitter(ctx, "featurecollection", NewFeatureCollectionEmitter)
}

// FeatureCollectionEmitter implements the `Emitter` interface for crawling features in a GeoJSON FeatureCollection record.
type FeatureCollectionEmitter struct {
	Emitter
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
}

// NewFeatureCollectionEmitter() returns a new `FeatureCollectionEmitter` instance configured by 'uri' in the form of:
//
//	featurecollection://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewFeatureCollectionEmitter(ctx context.Context, uri string) (Emitter, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	i := &FeatureCollectionEmitter{
		filters: f,
	}

	return i, nil
}

// WalkURI() walks (crawls) each feature in the GeoJSON FeatureCollection found in the file identified by 'uri' and for
// each file (not excluded by any filters specified when `idx` was created) invokes 'index_cb'.
func (idx *FeatureCollectionEmitter) WalkURI(ctx context.Context, index_cb EmitterCallbackFunc, uri string) error {

	fh, err := ReaderWithPath(ctx, uri)

	if err != nil {
		return fmt.Errorf("Failed to create reader for '%s', %w", uri, err)
	}

	defer fh.Close()

	body, err := io.ReadAll(fh)

	if err != nil {
		return fmt.Errorf("Failed to read body for '%s', %w", uri, err)
	}

	type FC struct {
		Type     string
		Features []interface{}
	}

	var collection FC

	err = json.Unmarshal(body, &collection)

	if err != nil {
		return fmt.Errorf("Failed to unmarshal '%s' as a feature collection, %w", uri, err)
	}

	for i, f := range collection.Features {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		path := fmt.Sprintf("%s#%d", uri, i)

		feature, err := json.Marshal(f)

		if err != nil {
			return fmt.Errorf("Failed to marshal feature for '%s', %w", path, err)
		}

		br := bytes.NewReader(feature)
		fh, err := ioutil.NewReadSeekCloser(br)

		if err != nil {
			return fmt.Errorf("Failed to create new ReadSeekCloser for '%s', %w", path, err)
		}

		if idx.filters != nil {

			ok, err := idx.filters.Apply(ctx, fh)

			if err != nil {
				return fmt.Errorf("Failed to apply filters for '%s', %w", path, err)
			}

			if !ok {
				continue
			}

			_, err = fh.Seek(0, 0)

			if err != nil {
				return fmt.Errorf("Failed to seek(0, 0) for '%s', %w", path, err)
			}
		}

		err = index_cb(ctx, path, fh)

		if err != nil {
			return fmt.Errorf("Index callback failed for '%s', %w", path, err)
		}
	}

	return nil
}
