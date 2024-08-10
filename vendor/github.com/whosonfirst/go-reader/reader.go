package reader

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"io"
	"net/url"
	"sort"
	"strings"
)

var reader_roster roster.Roster

// ReaderInitializationFunc is a function defined by individual reader package and used to create
// an instance of that reader
type ReaderInitializationFunc func(ctx context.Context, uri string) (Reader, error)

// Reader is an interface for reading data from multiple sources or targets.
type Reader interface {
	// Reader returns a `io.ReadSeekCloser` instance for a URI resolved by the instance implementing the `Reader` interface.
	Read(context.Context, string) (io.ReadSeekCloser, error)
	// The absolute path for the file is determined by the instance implementing the `Reader` interface.
	ReaderURI(context.Context, string) string
}

// RegisterReader registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Reader` instances by the `NewReader` method.
func RegisterReader(ctx context.Context, scheme string, init_func ReaderInitializationFunc) error {

	err := ensureReaderRoster()

	if err != nil {
		return err
	}

	return reader_roster.Register(ctx, scheme, init_func)
}

func ensureReaderRoster() error {

	if reader_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		reader_roster = r
	}

	return nil
}

// NewReader returns a new `Reader` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `ReaderInitializationFunc`
// function used to instantiate the new `Reader`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterReader` method.
func NewReader(ctx context.Context, uri string) (Reader, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := reader_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(ReaderInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureReaderRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range reader_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
