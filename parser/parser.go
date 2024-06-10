package parser

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

type Parser interface {
	Parse(context.Context, []byte) (*Components, error)
}

// ParserInitializationFunc is a function defined by individual parser package and used to create
// an instance of that parser
type ParserInitializationFunc func(ctx context.Context, uri string) (Parser, error)

var parser_roster roster.Roster

// RegisterParser registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Parser` instances by the `NewParser` method.
func RegisterParser(ctx context.Context, scheme string, init_func ParserInitializationFunc) error {

	err := ensureParserRoster()

	if err != nil {
		return err
	}

	return parser_roster.Register(ctx, scheme, init_func)
}

func ensureParserRoster() error {

	if parser_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		parser_roster = r
	}

	return nil
}

// NewParser returns a new `Parser` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `ParserInitializationFunc`
// function used to instantiate the new `Parser`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterParser` method.
func NewParser(ctx context.Context, uri string) (Parser, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := parser_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(ParserInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func ParserSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureParserRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range parser_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
