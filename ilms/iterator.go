package ilms

// > go run cmd/index-locations/main.go -verbose -location-database-uri null:// -location-parser-uri ilms:// -iterator-uri ilms:// /usr/local/data/ilms/*.csv

import (
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-dedupe/iterator"
)

type ILMSIterator struct {
	iterator.Iterator
	max_workers int
}

func init() {
	ctx := context.Background()
	err := iterator.RegisterIterator(ctx, "ilms", NewILMSIterator)
	if err != nil {
		panic(err)
	}
}

func NewILMSIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

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

	iter := &ILMSIterator{
		max_workers: max_workers,
	}

	return iter, nil
}

func (iter *ILMSIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[[]byte, error] {

	return func(yield func([]byte, error) bool) {

		/*
			throttle := make(chan bool, iter.max_workers)

			for i := 0; i < iter.max_workers; i++ {
				throttle <- true
			}
		*/

		done_ch := make(chan bool)
		err_ch := make(chan error)

		for _, path := range uris {

			go func(path string) {

				defer func() {
					done_ch <- true
				}()

				for body, err := range iter.iteratePath(ctx, path) {
					yield(body, err)
				}

			}(path)
		}

		remaining := len(uris)

		for remaining > 0 {
			select {
			case <-done_ch:
				remaining -= 1
			case err := <-err_ch:
				slog.Error(err.Error())
			}
		}

		return
	}
}

func (iter *ILMSIterator) Close(ctx context.Context) error {
	return nil
}

func (iter *ILMSIterator) iteratePath(ctx context.Context, path string) iter.Seq2[[]byte, error] {

	return func(yield func([]byte, error) bool) {

		csv_r, err := csvdict.NewReaderFromPath(path)

		if err != nil {
			yield(nil, fmt.Errorf("Failed to create CSV reader for %s, %w", path, err))
		}

		for {
			row, err := csv_r.Read()

			if err == io.EOF {
				break
			}

			if err != nil {
				yield(nil, err)
				return
			}

			logger := slog.Default()
			logger = logger.With("mid", row["MID"])
			logger = logger.With("name", row["COMMONNAME"])

			str_lat, ok := row["LATITUDE"]

			if !ok || strings.TrimSpace(str_lat) == "" {
				logger.Debug("Row is missing latitude, skipping")
				continue
			}

			str_lon, ok := row["LONGITUDE"]

			if !ok || strings.TrimSpace(str_lon) == "" {
				logger.Debug("Row is missing longitude, skipping")
				continue
			}

			lat, err := strconv.ParseFloat(str_lat, 64)

			if err != nil {
				logger.Warn("Invalid latitude for row, skipping", "latitude", str_lat, "error", err)
				continue
			}

			lon, err := strconv.ParseFloat(str_lon, 64)

			if err != nil {
				logger.Warn("Invalid longitude for row, skipping", "longitude", str_lon, "error", err)
				continue
			}

			pt := orb.Point([2]float64{lon, lat})

			f := geojson.NewFeature(pt)

			for k, v := range row {
				f.Properties[k] = v
			}

			enc_f, err := f.MarshalJSON()

			if err != nil {
				logger.Warn("Failed to marshal feature for row, skipping", "error", err)
				continue
			}

			if !yield(enc_f, nil) {
				logger.Warn("Callback failed for row", "error", err)
			}
		}

		return
	}
}
