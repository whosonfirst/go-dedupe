package uid

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"log"
	"net/url"
	"sort"
	"strings"
)

type Provider interface {
	UID(context.Context, ...interface{}) (UID, error)
	SetLogger(context.Context, *log.Logger) error
}

type UID interface {
	Value() any
	String() string
}

func AsInt64(u UID) (int64, bool) {
	v, ok := u.Value().(int64)
	return v, ok
}

var provider_roster roster.Roster

// ProviderInitializationFunc is a function defined by individual provider package and used to create
// an instance of that provider
type ProviderInitializationFunc func(ctx context.Context, uri string) (Provider, error)

// RegisterProvider registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Provider` instances by the `NewProvider` method.
func RegisterProvider(ctx context.Context, scheme string, init_func ProviderInitializationFunc) error {

	err := ensureProviderRoster()

	if err != nil {
		return err
	}

	return provider_roster.Register(ctx, scheme, init_func)
}

func ensureProviderRoster() error {

	if provider_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		provider_roster = r
	}

	return nil
}

// NewProvider returns a new `Provider` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `ProviderInitializationFunc`
// function used to instantiate the new `Provider`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterProvider` method.
func NewProvider(ctx context.Context, uri string) (Provider, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := provider_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(ProviderInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureProviderRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range provider_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
