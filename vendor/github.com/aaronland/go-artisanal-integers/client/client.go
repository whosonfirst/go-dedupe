package client

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"net/url"
	"sort"
	"strings"
)

type Client interface {
	NextInt(context.Context) (int64, error)
}

type ClientInitializeFunc func(ctx context.Context, uri string) (Client, error)

var clients roster.Roster

func ensureClientRoster() error {

	if clients == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		clients = r
	}

	return nil
}

func RegisterClient(ctx context.Context, scheme string, f ClientInitializeFunc) error {

	err := ensureClientRoster()

	if err != nil {
		return err
	}

	return clients.Register(ctx, scheme, f)
}

func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureClientRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range clients.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

func NewClient(ctx context.Context, uri string) (Client, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := clients.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	f := i.(ClientInitializeFunc)
	return f(ctx, uri)
}
