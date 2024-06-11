// Package emitter provides an interface for crawling data sources and "emitting" records.
package emitter

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"
)

// STDIN is a constant value signaling that a record was read from `STDIN` and has no URI (path).
const STDIN string = "STDIN"

// type EmitterInitializeFunc is a function used to initialize an implementation of the `Emitter` interface.
type EmitterInitializeFunc func(context.Context, string) (Emitter, error)

// EmitterCallbackFunc is a custom function used to process individual records as they are crawled by an instance of the `Emitter` interface.
type EmitterCallbackFunc func(context.Context, string, io.ReadSeeker, ...interface{}) error

// type Emitter is an interface for crawling data sources and "emitting" records. Data sources are assumed to be Who's On First records.
type Emitter interface {
	WalkURI(context.Context, EmitterCallbackFunc, string) error
}

// emitters is a `aaronland/go-roster.Roster` instance used to maintain a list of registered `EmitterInitializeFunc` initialization functions.
var emitters roster.Roster

// RegisterEmitter() associates 'scheme' with 'init_func' in an internal list of avilable `Emitter` implementations.
func RegisterEmitter(ctx context.Context, scheme string, f EmitterInitializeFunc) error {

	err := ensureSpatialRoster()

	if err != nil {
		return fmt.Errorf("Failed to register %s scheme, %w", scheme, err)
	}

	return emitters.Register(ctx, scheme, f)
}

// NewEmitter() returns a new `Emitter` instance derived from 'uri'. The semantics of and requirements for
// 'uri' as specific to the package implementing the interface.
func NewEmitter(ctx context.Context, uri string) (Emitter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	scheme := u.Scheme

	if scheme == "" {
		return nil, fmt.Errorf("Emittter URI is missing scheme '%s'", uri)
	}

	i, err := emitters.Driver(ctx, scheme)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve driver for '%s' scheme, %w", scheme, err)
	}

	fn := i.(EmitterInitializeFunc)

	if fn == nil {
		return nil, fmt.Errorf("Unregistered initialization function for '%s' scheme", scheme)
	}

	return fn(ctx, uri)
}

// Schemes() returns the list of schemes that have been "registered".
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSpatialRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range emitters.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

// ReaderWithPath returns a new `io.ReadSeekCloser` instance derived from 'abs_path'.
func ReaderWithPath(ctx context.Context, abs_path string) (io.ReadSeekCloser, error) {

	if abs_path == STDIN {
		return os.Stdin, nil
	}

	fh, err := os.Open(abs_path)

	if err != nil {
		return nil, fmt.Errorf("Failed to open %s, %w", abs_path, err)
	}

	return fh, nil
}

// ensureDispatcherRoster() ensures that a `aaronland/go-roster.Roster` instance used to maintain a list of registered `EmitterInitializeFunc`
// initialization functions is present
func ensureSpatialRoster() error {

	if emitters == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return fmt.Errorf("Failed to create new roster, %w", err)
		}

		emitters = r
	}

	return nil
}
