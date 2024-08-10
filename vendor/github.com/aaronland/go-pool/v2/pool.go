package pool

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"net/url"
	"sort"
	"strings"
)

type Pool interface {
	Length(context.Context) int64
	Push(context.Context, any) error
	Pop(context.Context) (any, bool)
}

var pool_roster roster.Roster

// PoolInitializationFunc is a function defined by individual pool package and used to create
// an instance of that pool
type PoolInitializationFunc func(ctx context.Context, uri string) (Pool, error)

// RegisterPool registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Pool` instances by the `NewPool` method.
func RegisterPool(ctx context.Context, scheme string, init_func PoolInitializationFunc) error {

	err := ensurePoolRoster()

	if err != nil {
		return err
	}

	return pool_roster.Register(ctx, scheme, init_func)
}

func ensurePoolRoster() error {

	if pool_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		pool_roster = r
	}

	return nil
}

// NewPool returns a new `Pool` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `PoolInitializationFunc`
// function used to instantiate the new `Pool`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterPool` method.
func NewPool(ctx context.Context, uri string) (Pool, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := pool_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(PoolInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensurePoolRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range pool_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

