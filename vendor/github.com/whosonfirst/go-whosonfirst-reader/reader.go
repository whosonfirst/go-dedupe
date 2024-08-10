package reader

import (
	"context"
	"fmt"
	"io"

	"github.com/paulmach/orb/geojson"
	go_reader "github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

// LoadReadCloser will return an `io.ReadCloser` instance from 'r' for the relative path associated with 'id'.
func LoadReadCloser(ctx context.Context, r go_reader.Reader, id int64) (io.ReadCloser, error) {

	rel_path, err := uri.Id2RelPath(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive path for %d, %w", id, err)
	}

	rc, err := r.Read(ctx, rel_path)

	if err != nil {
		return nil, fmt.Errorf("Failed to read %s, %w", rel_path, err)
	}

	return rc, nil
}

// LoadBytes will return an `[]byte` instance from 'r' for the relative path associated with 'id'.
func LoadBytes(ctx context.Context, r go_reader.Reader, id int64) ([]byte, error) {

	fh, err := LoadReadCloser(ctx, r, id)

	if err != nil {
		return nil, fmt.Errorf("Failed to load handle for %d, %w", id, err)
	}

	defer fh.Close()

	body, err := io.ReadAll(fh)

	if err != nil {
		return nil, fmt.Errorf("Failed to read data for %d, %w", id, err)
	}

	return body, nil
}

// LoadBytes will return an `paulmach/orb/geojson.Feature` instance from 'r' for the relative path associated with 'id'.
func LoadFeature(ctx context.Context, r go_reader.Reader, id int64) (*geojson.Feature, error) {

	body, err := LoadBytes(ctx, r, id)

	if err != nil {
		return nil, fmt.Errorf("Failed to load bytes from %d, %w", id, err)
	}

	f, err := geojson.UnmarshalFeature(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal GeoJSON from %d, %w", id, err)
	}

	return f, nil
}
