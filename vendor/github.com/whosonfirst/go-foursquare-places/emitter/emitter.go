package emitter

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
	"github.com/whosonfirst/go-foursquare-places"
)

type Emitter interface {
	Emit(context.Context) iter.Seq2[*places.Place, error]
	Close() error
}

var emitter_roster roster.Roster

// EmitterInitializationFunc is a function defined by individual emitter package and used to create
// an instance of that emitter
type EmitterInitializationFunc func(ctx context.Context, uri string) (Emitter, error)

// RegisterEmitter registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Emitter` instances by the `NewEmitter` method.
func RegisterEmitter(ctx context.Context, scheme string, init_func EmitterInitializationFunc) error {

	err := ensureEmitterRoster()

	if err != nil {
		return err
	}

	return emitter_roster.Register(ctx, scheme, init_func)
}

func ensureEmitterRoster() error {

	if emitter_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		emitter_roster = r
	}

	return nil
}

// NewEmitter returns a new `Emitter` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `EmitterInitializationFunc`
// function used to instantiate the new `Emitter`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterEmitter` method.
func NewEmitter(ctx context.Context, uri string) (Emitter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := emitter_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(EmitterInitializationFunc)
	return init_func(ctx, uri)
}

// EmitterSchemes returns the list of schemes that have been registered.
func EmitterSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureEmitterRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range emitter_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
