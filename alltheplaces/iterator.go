package alltheplaces

import (
	"context"
	"fmt"

	"github.com/whosonfirst/go-dedupe/iterator"
)

type AllThePlacesIterator struct {
	iterator.Iterator
}

func init() {
	ctx := context.Background()
	err := iterator.RegisterIterator(ctx, "alltheplaces", NewAllThePlacesIterator)
	if err != nil {
		panic(err)
	}
}

func NewAllThePlacesIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

	i := &AllThePlacesIterator{}
	return i, nil
}

func (i *AllThePlacesIterator) IterateWithCallback(ctx context.Context, cb iterator.IteratorCallback, uris ...string) error {

	return fmt.Errorf("Not implemeted")
}
