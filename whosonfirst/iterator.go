package whosonfirst

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/whosonfirst/go-dedupe/iterator"
	wof_iterator "github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

type WhosOnFirstIterator struct {
	iterator.Iterator
	iterator_uri string
}

func init() {
	ctx := context.Background()
	err := iterator.RegisterIterator(ctx, "whosonfirst", NewWhosOnFirstIterator)
	if err != nil {
		panic(err)
	}
}

func NewWhosOnFirstIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()
	iter_uri := q.Get("iterator-uri")

	i := &WhosOnFirstIterator{
		iterator_uri: iter_uri,
	}

	return i, nil
}

func (i *WhosOnFirstIterator) IterateWithCallback(ctx context.Context, cb iterator.IteratorCallback, uris ...string) error {

	wof_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read both for %s, %w", path, err)
		}

		return cb(ctx, body)
	}

	wof_iter, err := wof_iterator.NewIterator(ctx, i.iterator_uri, wof_cb)

	if err != nil {
		return err
	}

	return wof_iter.IterateURIs(ctx, uris...)
}

func (iter *WhosOnFirstIterator) Close(ctx context.Context) error {
	return nil
}
