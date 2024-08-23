package whosonfirst

import (
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"
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

	if iter_uri == "" {
		iter_uri = "repo://?exclude=properties.edtf:deprecated=.*"
		slog.Debug("No WOF iterator URI defined, assigning default URI", "uri", iter_uri)
	}

	i := &WhosOnFirstIterator{
		iterator_uri: iter_uri,
	}

	return i, nil
}

func (i *WhosOnFirstIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[[]byte, error] {

	return func(yield func([]byte, error) bool) {

		wof_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

			body, err := io.ReadAll(r)

			if err != nil {
				yield(nil, err)
				return fmt.Errorf("Failed to read both for %s, %w", path, err)
			}

			if !yield(body, nil) {
				return fmt.Errorf("Failed to yield record for %s", path)
			}

			return nil
		}

		wof_iter, err := wof_iterator.NewIterator(ctx, i.iterator_uri, wof_cb)

		if err != nil {
			yield(nil, err)
			return
		}

		err = wof_iter.IterateURIs(ctx, uris...)

		if err != nil {
			yield(nil, err)
			return
		}

		return
	}
}

func (iter *WhosOnFirstIterator) Close(ctx context.Context) error {
	return nil
}
