package overture

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"sync"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/whosonfirst/go-dedupe/iterator"
	"github.com/whosonfirst/go-overture/geojsonl"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
)

type OvertureIterator struct {
	iterator.Iterator
	bucket      *blob.Bucket
	is_bzipped  bool
	max_workers int
	start_after int
}

func init() {
	ctx := context.Background()
	err := iterator.RegisterIterator(ctx, "overture", NewOvertureIterator)
	if err != nil {
		panic(err)
	}
}

func NewOvertureIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	bucket_uri := q.Get("bucket-uri")

	source_bucket, err := bucket.OpenBucket(ctx, bucket_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to open bucket, %w", err)
	}

	max_workers := 20

	if q.Has("max-workers") {

		v, err := strconv.Atoi(q.Get("max-workers"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?max-workers= parameter, %w", err)
		}

		max_workers = v
	}

	is_bzipped := true

	if q.Has("is-bzipped") {

		v, err := strconv.ParseBool(q.Get("is-bzipped"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?is-bzipped= parameter, %w", err)
		}

		is_bzipped = v
	}

	start_after := 0

	if q.Has("start-after") {

		v, err := strconv.Atoi(q.Get("start-after"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?start-after= parameter, %w", err)
		}

		start_after = v
	}

	iter := &OvertureIterator{
		bucket:      source_bucket,
		max_workers: max_workers,
		is_bzipped:  is_bzipped,
		start_after: start_after,
	}

	return iter, nil
}

func (iter *OvertureIterator) IterateWithCallback(ctx context.Context, cb iterator.IteratorCallback, uris ...string) error {

	throttle := make(chan bool, iter.max_workers)

	for i := 0; i < iter.max_workers; i++ {
		throttle <- true
	}

	wg := new(sync.WaitGroup)

	walk_cb := func(ctx context.Context, path string, rec *walk.WalkRecord) error {

		if iter.start_after > 0 && rec.LineNumber < iter.start_after {
			// monitor.Signal(ctx)
			return nil
		}

		<-throttle

		wg.Add(1)

		go func(path string, rec *walk.WalkRecord) {

			logger := slog.Default()
			logger = logger.With("path", path)
			logger = logger.With("line number", rec.LineNumber)

			defer func() {
				wg.Done()
				throttle <- true
			}()

			err := cb(ctx, rec.Body)

			if err != nil {
				logger.Error("Iterator callback for record failed", "error", err)
			}

		}(path, rec)

		return nil
	}

	walk_opts := &geojsonl.WalkOptions{
		SourceBucket: iter.bucket,
		Callback:     walk_cb,
		IsBzipped:    iter.is_bzipped,
	}

	err := geojsonl.Walk(ctx, walk_opts, uris...)

	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}

func (iter *OvertureIterator) Close(ctx context.Context) error {
	return iter.bucket.Close()
}
