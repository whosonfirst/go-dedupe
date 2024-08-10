package export

import (
	"context"
	"net/url"

	"github.com/aaronland/go-roster"
)

type Exporter interface {
	Export(context.Context, []byte) ([]byte, error)
	ExportFeature(context.Context, interface{}) ([]byte, error)
}

var exporter_roster roster.Roster

type ExporterInitializationFunc func(ctx context.Context, uri string) (Exporter, error)

func RegisterExporter(ctx context.Context, scheme string, init_func ExporterInitializationFunc) error {

	err := ensureExporterRoster()

	if err != nil {
		return err
	}

	return exporter_roster.Register(ctx, scheme, init_func)
}

func ensureExporterRoster() error {

	if exporter_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		exporter_roster = r
	}

	return nil
}

func NewExporter(ctx context.Context, uri string) (Exporter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := exporter_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(ExporterInitializationFunc)
	return init_func(ctx, uri)
}

func Exporters() []string {
	ctx := context.Background()
	return exporter_roster.Drivers(ctx)
}
