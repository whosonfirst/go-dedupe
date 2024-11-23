package foursquare

// ./bin/index-locations -iterator-uri 'foursquare://?emitter-uri=csv:///usr/local/data/4sq/4sq.csv.bz2' -location-parser-uri 'foursquare://' -location-database-uri 'sql://sqlite3?dsn=/usr/local/data/4sq/4sq-locations.db'

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/whosonfirst/go-dedupe/iterator"
	"github.com/whosonfirst/go-foursquare-places/emitter"
)

type FoursquareIterator struct {
	iterator.Iterator
	emitter emitter.Emitter
}

func init() {
	ctx := context.Background()
	err := iterator.RegisterIterator(ctx, "foursquare", NewFoursquareIterator)
	if err != nil {
		panic(err)
	}
}

func NewFoursquareIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	emitter_uri := q.Get("emitter-uri")

	e, err := emitter.NewEmitter(ctx, emitter_uri)

	if err != nil {
		return nil, err
	}

	iter := &FoursquareIterator{
		emitter: e,
	}

	return iter, nil
}

func (iter *FoursquareIterator) IterateWithCallback(ctx context.Context, cb iterator.IteratorCallback, uris ...string) error {

	var iter_err error

	for pl, err := range iter.emitter.Emit(ctx) {

		if err != nil {
			iter_err = fmt.Errorf("Failed to iterate places, %w", err)
			break
		}

		body, err := json.Marshal(pl)

		if err != nil {
			iter_err = fmt.Errorf("Failed to marshal place %s, %w", pl, err)
			break
		}

		err = cb(ctx, body)

		if err != nil {
			iter_err = fmt.Errorf("Failed to execute callback for place %s, %w", pl, err)
			break
		}
	}

	if iter_err != nil {
		return iter_err
	}

	return nil
}

func (iter *FoursquareIterator) Close(ctx context.Context) error {
	return iter.emitter.Close()
}
