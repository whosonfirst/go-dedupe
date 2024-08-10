package export

import (
	"context"

	id "github.com/whosonfirst/go-whosonfirst-id"
)

type Options struct {
	IDProvider id.Provider
}

func NewDefaultOptions(ctx context.Context) (*Options, error) {

	provider, err := id.NewProvider(ctx)

	if err != nil {
		return nil, err
	}

	return NewDefaultOptionsWithProvider(ctx, provider)
}

func NewDefaultOptionsWithProvider(ctx context.Context, provider id.Provider) (*Options, error) {

	opts := &Options{
		IDProvider: provider,
	}

	return opts, nil
}
