package alltheplaces

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
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

	max_workers := 20

	iter := &AllThePlacesIterator{
		max_workers: max_workers,
	}

	return iter, nil
}

func (iter *AllThePlacesIterator) IterateWithCallback(ctx context.Context, cb iterator.IteratorCallback, uris ...string) error {

	throttle := make(chan bool, iter.max_workers)

	for i := 0; i < iter.max_workers; i++ {
		throttle <- true
	}

	wg := new(sync.WaitGroup)

	for _, path := range uris {

		r, err := os.Open(path)

		if err != nil {
			return fmt.Errorf("Failed to open %s for reading, %v", path, err)
		}

		defer r.Close()

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %v", path, err)
		}

		slog.Info("Unmarshal", "path", path)

		fc, err := geojson.UnmarshalFeatureCollection(body)

		if err != nil {
			slog.Warn("Failed to unmarshal feature collection", "path", path, "error", err)
			continue
		}

		for _, f := range fc.Features {

			<-throttle

			wg.Add(1)

			go func(f *geojson.Feature) {

				defer func() {
					wg.Done()
					throttle <- true
				}()

				body, err := f.MarshalJSON()

				if err != nil {
					slog.Error("Failed to marshal record", "error", err)
					return
				}

				err = cb(ctx, body)

				if err != nil {
					slog.Error("Callback failed for record", "error", err)
					return
				}
			}(f)
		}
	}

	wg.Done()

	return nil
}
