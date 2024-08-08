package overture

import (
	"context"
	// "fmt"
	"log/slog"
	"sync"

	"github.com/aaronland/go-jsonl/walk"
	// "github.com/aaronland/gocloud-blob/bucket"
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
}

func init() {
	ctx := context.Background()
	err := iterator.RegisterIterator(ctx, "overture", NewOvertureIterator)
	if err != nil {
		panic(err)
	}
}

func NewOvertureIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

	max_workers := 20

	iter := &OvertureIterator{
		max_workers: max_workers,
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

		/*
			if start_after > 0 && rec.LineNumber < start_after {
				monitor.Signal(ctx)
				return nil
			}
		*/

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
