package iterator

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

type IteratorCallback func(context.Context, []byte) error

// Iterator is an interface for procesing arbitrary data sources that yield individual JSON-encoded GeoJSON Features.
type Iterator interface {
	// Iterate(context.Context, ...string) iter.Seq2[*geojson.Feature, error]
	IterateWithCallback(context.Context, IteratorCallback, ...string) error
	// Close performs and terminating functions required by the iterator
	Close(context.Context) error
}

// IteratorInitializationFunc is a function defined by individual iterator package and used to create
// an instance of that iterator
type IteratorInitializationFunc func(ctx context.Context, uri string) (Iterator, error)

var iterator_roster roster.Roster

// RegisterIterator registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Iterator` instances by the `NewIterator` method.
func RegisterIterator(ctx context.Context, scheme string, init_func IteratorInitializationFunc) error {

	err := ensureIteratorRoster()

	if err != nil {
		return err
	}

	return iterator_roster.Register(ctx, scheme, init_func)
}

func ensureIteratorRoster() error {

	if iterator_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		iterator_roster = r
	}

	return nil
}

// NewIterator returns a new `Iterator` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `IteratorInitializationFunc`
// function used to instantiate the new `Iterator`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterIterator` method.
func NewIterator(ctx context.Context, uri string) (Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := iterator_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(IteratorInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func IteratorSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureIteratorRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range iterator_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
