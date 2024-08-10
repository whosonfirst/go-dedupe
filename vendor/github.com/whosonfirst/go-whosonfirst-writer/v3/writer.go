package writer

import (
	"bytes"
	"context"
	"fmt"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-uri"
	go_writer "github.com/whosonfirst/go-writer/v3"
)

// WriteFeature will serialize and write 'f' using 'wr' using a default `whosonfirst/go-whosonfirst-export/v2.Exporter` instance.
func WriteFeature(ctx context.Context, wr go_writer.Writer, f *geojson.Feature) (int64, error) {

	body, err := f.MarshalJSON()

	if err != nil {
		return -1, fmt.Errorf("Failed to marshal JSON, %w", err)
	}

	return WriteBytes(ctx, wr, body)
}

// WriteFeatureWithExporter will serialize and write 'f' using 'wr' using a custom `whosonfirst/go-whosonfirst-export/v2.Exporter` instance.
func WriteFeatureWithExporter(ctx context.Context, wr go_writer.Writer, ex export.Exporter, f *geojson.Feature) (int64, error) {

	body, err := f.MarshalJSON()

	if err != nil {
		return -1, fmt.Errorf("Failed to marshal JSON, %w", err)
	}

	return WriteBytesWithExporter(ctx, wr, ex, body)
}

// WriteBytes will write 'body' using 'wr' using a default `whosonfirst/go-whosonfirst-export/v2.Exporter` instance.
func WriteBytes(ctx context.Context, wr go_writer.Writer, body []byte) (int64, error) {

	ex, err := export.NewExporter(ctx, "whosonfirst://")

	if err != nil {
		return -1, fmt.Errorf("Failed to create new exporter, %w", err)
	}

	return WriteBytesWithExporter(ctx, wr, ex, body)
}

// WriteBytesWithExporter will write 'body' using 'wr' using a custom `whosonfirst/go-whosonfirst-export/v2.Exporter` instance.
func WriteBytesWithExporter(ctx context.Context, wr go_writer.Writer, ex export.Exporter, body []byte) (int64, error) {

	body, err := ex.Export(ctx, body)

	if err != nil {
		return -1, fmt.Errorf("Failed to export data, %w", err)
	}

	id, err := properties.Id(body)

	if err != nil {
		return -1, fmt.Errorf("Failed to derive ID, %w", err)
	}

	// START OF put me in a function somewhere...

	var rel_path string

	if alt.IsAlt(body) {

		alt_label, err := properties.AltLabel(body)

		if err != nil {
			return -1, fmt.Errorf("Failed to derive alt label, %w", err)
		}

		uri_args, err := uri.NewAlternateURIArgsFromAltLabel(alt_label)

		if err != nil {
			return -1, fmt.Errorf("Failed to derive URI args from label '%s', %w", alt_label, err)
		}

		rel_path, err = uri.Id2RelPath(id, uri_args)

	} else {
		rel_path, err = uri.Id2RelPath(id)
	}

	if err != nil {
		return -1, fmt.Errorf("Failed to derive relative path, %w", err)
	}

	// END OF put me in a function somewhere...

	br := bytes.NewReader(body)

	_, err = wr.Write(ctx, rel_path, br)

	if err != nil {
		return -1, fmt.Errorf("Failed to write %s, %w", rel_path, err)
	}

	return id, nil
}
