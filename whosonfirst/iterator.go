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

		slog.Debug("OMGWTF", "yield", yield)
		wof_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

			slog.Debug("BBQ", "path", path)
			body, err := io.ReadAll(r)

			if err != nil {
				slog.Error("WUT", "error", err)
				yield(nil, err)
				return fmt.Errorf("Failed to read both for %s, %w", path, err)
			}

			slog.Debug("YIELD", "body", len(body))
			ok := yield(nil, nil)
			slog.Debug("OK", "ok", ok)
			return nil
		}

		wof_iter, err := wof_iterator.NewIterator(ctx, i.iterator_uri, wof_cb)

		if err != nil {
			slog.Debug("ERROR1", "error", err)
			// yield(nil, err)
			return
		}

		err = wof_iter.IterateURIs(ctx, uris...)

		if err != nil {
			slog.Debug("ERROR2", "error", err)			
			// yield(nil, err)
			return
		}

		return
	}
}

func (iter *WhosOnFirstIterator) Close(ctx context.Context) error {
	return nil
}
