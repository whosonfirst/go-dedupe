package overture

import (
	"context"
	"fmt"

	"github.com/whosonfirst/go-dedupe/iterator"
)

type OvertureIterator struct {
	iterator.Iterator
}

func init() {
	ctx := context.Background()
	err := iterator.RegisterIterator(ctx, "overture", NewOvertureIterator)
	if err != nil {
		panic(err)
	}
}

func NewOvertureIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

	i := &OvertureIterator{}
	return i, nil
}

func (i *OvertureIterator) IterateWithCallback(ctx context.Context, cb iterator.IteratorCallback, uris ...string) error {

	return fmt.Errorf("Not implemeted")
}
