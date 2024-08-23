package alltheplaces

// > go run cmd/index-locations/main.go -verbose -location-database-uri null:// -location-parser-uri alltheplaces:// -iterator-uri alltheplaces:// /usr/local/data/alltheplaces/*.geojson

import (
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"sync"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-dedupe/iterator"
)

type AllThePlacesIterator struct {
	iterator.Iterator
	max_workers int
}

func init() {
	ctx := context.Background()
	err := iterator.RegisterIterator(ctx, "alltheplaces", NewAllThePlacesIterator)
	if err != nil {
		panic(err)
	}
}

func NewAllThePlacesIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	max_workers := 20

	if q.Has("max-workers") {

		v, err := strconv.Atoi(q.Get("max-workers"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?max-workers= parameter, %w", err)
		}

		max_workers = v
	}

	iter := &AllThePlacesIterator{
		max_workers: max_workers,
	}

	return iter, nil
}

func (iter *AllThePlacesIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[[]byte, error] {

	return func(yield func([]byte, error) bool) {

		throttle := make(chan bool, iter.max_workers)

		for i := 0; i < iter.max_workers; i++ {
			throttle <- true
		}

		wg := new(sync.WaitGroup)

		for _, path := range uris {

			logger := slog.Default()
			logger = logger.With("path", path)

			logger.Debug("Process record")

			r, err := os.Open(path)

			if err != nil {
				yield(nil, fmt.Errorf("Failed to open %s for reading, %v", path, err))
				return
			}

			defer r.Close()

			body, err := io.ReadAll(r)

			if err != nil {
				yield(nil, fmt.Errorf("Failed to read %s, %v", path, err))
				return
			}

			fc, err := geojson.UnmarshalFeatureCollection(body)

			if err != nil {
				logger.Warn("Failed to unmarshal feature collection", "path", path, "error", err)
				continue
			}

			for offset, f := range fc.Features {

				/*
				<-throttle

				wg.Add(1)

				 go func(offset int, f *geojson.Feature) {

					defer func() {
						wg.Done()
						throttle <- true
					}()
				*/
				
					body, err := f.MarshalJSON()

					if err != nil {
						yield(nil, fmt.Errorf("Failed to marshal record", "offset", offset, "error", err))
						return
					}

					yield(body, nil)

				// }(offset, f)
			}
		}

		wg.Wait()
		return
	}
}

func (iter *AllThePlacesIterator) Close(ctx context.Context) error {
	return nil
}
