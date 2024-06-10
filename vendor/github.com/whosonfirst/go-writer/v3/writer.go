package writer

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"io"
	"log"
	"net/url"
	"sort"
	"strings"
)

var writer_roster roster.Roster

// WriterInitializationFunc is a function defined by individual writer package and used to create
// an instance of that writer
type WriterInitializationFunc func(ctx context.Context, uri string) (Writer, error)

// Clone returns a new `Writer` instance derived from existing any new parameters.
// Clone(context.Context, string) (Writer, error)

// Writer is an interface for writing data to multiple sources or targets.
type Writer interface {
	// Writer copies the contents of an `io.ReadSeeker` instance to a relative path.
	// The absolute path for the file is determined by the instance implementing the `Writer` interface.
	Write(context.Context, string, io.ReadSeeker) (int64, error)
	// WriterURI returns the absolute URI for an instance implementing the `Writer` interface.
	WriterURI(context.Context, string) string
	// Flush publishes any outstanding data. The details of if, how or where data is "published" is determined by individual implementations.
	Flush(context.Context) error
	// Close closes any underlying writing mechnisms for an instance implementing the `Writer` interface.
	Close(context.Context) error
	// SetLogger assigns a custom logger to a `Writer` instance
	SetLogger(context.Context, *log.Logger) error
}

// RegisterWriter registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Writer` instances by the `NewWriter` method.
func RegisterWriter(ctx context.Context, scheme string, init_func WriterInitializationFunc) error {

	err := ensureWriterRoster()

	if err != nil {
		return err
	}

	return writer_roster.Register(ctx, scheme, init_func)
}

func ensureWriterRoster() error {

	if writer_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		writer_roster = r
	}

	return nil
}

// NewWriter returns a new `Writer` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `WriterInitializationFunc`
// function used to instantiate the new `Writer`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterWriter` method.
func NewWriter(ctx context.Context, uri string) (Writer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := writer_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(WriterInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureWriterRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range writer_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
