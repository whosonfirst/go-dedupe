package embeddings

// https://github.com/asg017/sqlite-vss

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

// Embedder defines an interface for generating (vector) embeddings
type Embedder interface {
	// Embeddings returns the embeddings for a string as a list of float64 values.
	Embeddings(context.Context, string) ([]float64, error)
	// Embeddings32 returns the embeddings for a string as a list of float32 values.
	Embeddings32(context.Context, string) ([]float32, error)
	// ImageEmbeddings returns the embeddings for a base64-encoded image as a list of float64 values.
	ImageEmbeddings(context.Context, []byte) ([]float64, error)
	// ImageEmbeddings32 returns the embeddings for a base64-encoded image as a list of float32 values.
	ImageEmbeddings32(context.Context, []byte) ([]float32, error)
}

// EmbedderInitializationFunc is a function defined by individual embedder package and used to create
// an instance of that embedder
type EmbedderInitializationFunc func(ctx context.Context, uri string) (Embedder, error)

var embedder_roster roster.Roster

// RegisterEmbedder registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Embedder` instances by the `NewEmbedder` method.
func RegisterEmbedder(ctx context.Context, scheme string, init_func EmbedderInitializationFunc) error {

	err := ensureEmbedderRoster()

	if err != nil {
		return err
	}

	return embedder_roster.Register(ctx, scheme, init_func)
}

func ensureEmbedderRoster() error {

	if embedder_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		embedder_roster = r
	}

	return nil
}

// NewEmbedder returns a new `Embedder` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `EmbedderInitializationFunc`
// function used to instantiate the new `Embedder`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterEmbedder` method.
func NewEmbedder(ctx context.Context, uri string) (Embedder, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := embedder_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(EmbedderInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func EmbedderSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureEmbedderRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range embedder_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
