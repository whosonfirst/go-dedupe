package export

import (
	"context"
	"encoding/json"
	"net/url"
)

type WhosOnFirstExporter struct {
	Exporter
	options *Options
}

func init() {

	ctx := context.Background()

	err := RegisterExporter(ctx, "whosonfirst", NewWhosOnFirstExporter)

	if err != nil {
		panic(err)
	}
}

func NewWhosOnFirstExporter(ctx context.Context, uri string) (Exporter, error) {

	_, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	opts, err := NewDefaultOptions(ctx)

	if err != nil {
		return nil, err
	}

	ex := WhosOnFirstExporter{
		options: opts,
	}

	return &ex, nil
}

func (ex *WhosOnFirstExporter) ExportFeature(ctx context.Context, feature interface{}) ([]byte, error) {

	body, err := json.Marshal(feature)

	if err != nil {
		return nil, err
	}

	return ex.Export(ctx, body)
}

func (ex *WhosOnFirstExporter) Export(ctx context.Context, feature []byte) ([]byte, error) {

	var err error

	feature, err = Prepare(feature, ex.options)

	if err != nil {
		return nil, err
	}

	feature, err = Format(feature, ex.options)

	if err != nil {
		return nil, err
	}

	return feature, nil
}
