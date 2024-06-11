// Package iterator provides methods and utilities for iterating over a collection of records
// (presumed but not required to be Who's On First records) from a variety of sources and dispatching
// processing to user-defined callback functions.
package iterator

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
)

// type Iterator provides a struct that can be used for iterating over a collection of records
// (presumed but not required to be Who's On First records) from a variety of sources and dispatching
// processing to user-defined callback functions.
type Iterator struct {
	// Emitter is a `emitter.Emitter` instance used to crawl and emit records.
	Emitter emitter.Emitter
	// EmitterCallbackFunc is a `emitter.EmitterCallbackFunc` callback function to be applied to each record emitted by `Emitter`.
	EmitterCallbackFunc emitter.EmitterCallbackFunc
	// Logger is a `log.Logger` instance used to provide feedback.
	Logger *log.Logger
	// Seen is the count of documents that have been seen (or emitted).
	Seen int64
	// count is the current number of documents being processed used to signal where an `Iterator` instance is still indexing (processing) documents.
	count int64
	// max_procs is the number maximum (CPU) processes to used to process documents simultaneously.
	max_procs int
	// exclude_paths is a `regexp.Regexp` instance used to test and exclude (if matching) the paths of documents as they are iterated through.
	exclude_paths *regexp.Regexp

	max_attempts int
	retry_after  int
}

// NewIterator() returns a new `Iterator` instance derived from 'emitter_uri' and 'emitter_cb'. The former is expected
// to be a valid `whosonfirst/go-whosonfirst-iterate/v2/emitter.Emitter` URI whose semantics are defined by the underlying
// implementation of the `emitter.Emitter` interface. The following iterator-specific query parameters are also accepted:
// * `?_max_procs=` Explicitly set the number maximum processes to use for iterating documents simultaneously. (Default is the value of `runtime.NumCPU()`.)
// * `?_exclude=` A valid regular expresion used to test and exclude (if matching) the paths of documents as they are iterated through.
func NewIterator(ctx context.Context, emitter_uri string, emitter_cb emitter.EmitterCallbackFunc) (*Iterator, error) {

	idx, err := emitter.NewEmitter(ctx, emitter_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new emitter, %w", err)
	}

	// Technically a no-op since we'll have parse 'emitter_uri' in NewEmitter but best not to assume

	u, err := url.Parse(emitter_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	max_procs := runtime.NumCPU()

	retry := false
	max_attempts := 1
	retry_after := 10 // seconds

	if q.Has("_max_procs") {

		max, err := strconv.ParseInt(q.Get("_max_procs"), 10, 64)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_max_procs' parameter, %w", err)
		}

		max_procs = int(max)
	}

	if q.Has("_retry") {

		v, err := strconv.ParseBool(q.Get("_retry"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_retry' parameter, %w", err)
		}

		retry = v
	}

	if retry {

		if q.Has("_max_retries") {

			v, err := strconv.Atoi(q.Get("_max_retries"))

			if err != nil {
				return nil, fmt.Errorf("Failed to parse '_max_retries' parameter, %w", err)
			}

			max_attempts = v
		}

		if q.Has("_retry_after") {

			v, err := strconv.Atoi(q.Get("_retry_after"))

			if err != nil {
				return nil, fmt.Errorf("Failed to parse '_retry_after' parameter, %w", err)
			}

			retry_after = v
		}
	}

	slog_logger := slog.Default()
	logger := slog.NewLogLogger(slog_logger.Handler(), slog.LevelInfo)

	i := Iterator{
		Emitter:             idx,
		EmitterCallbackFunc: emitter_cb,
		Logger:              logger,
		Seen:                0,
		count:               0,
		max_procs:           max_procs,
		max_attempts:        max_attempts,
		retry_after:         retry_after,
	}

	if q.Has("_exclude") {

		re_exclude, err := regexp.Compile(q.Get("_exclude"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_exclude' parameter, %w", err)
		}

		i.exclude_paths = re_exclude
	}

	return &i, nil
}

// IterateURIs processes 'uris' concurrent dispatching each URI to the iterator's underlying `Emitter.WalkURI`
// method and `EmitterCallbackFunc` callback function.
func (idx *Iterator) IterateURIs(ctx context.Context, uris ...string) error {

	t1 := time.Now()

	defer func() {
		t2 := time.Since(t1)
		idx.Logger.Printf("time to index paths (%d) %v", len(uris), t2)
	}()

	idx.increment()
	defer idx.decrement()

	local_callback := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		defer atomic.AddInt64(&idx.Seen, 1)

		if idx.exclude_paths != nil {

			if idx.exclude_paths.MatchString(path) {
				return nil
			}
		}

		return idx.EmitterCallbackFunc(ctx, path, fh, args...)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	procs := idx.max_procs
	throttle := make(chan bool, procs)

	for i := 0; i < procs; i++ {
		throttle <- true
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)

	remaining := len(uris)

	for _, uri := range uris {

		go func(uri string) {

			<-throttle

			defer func() {
				throttle <- true
				done_ch <- true
			}()

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			var walk_err error
			attempts := 0

			// slog.Info("Walk", "uri", uri, "max_attempts", idx.max_attempts, "retry after", idx.retry_after)

			for attempts < idx.max_attempts {

				// slog.Info("Walk URI", "uri", uri, "attempts", attempts)
				walk_err = idx.Emitter.WalkURI(ctx, local_callback, uri)

				if walk_err == nil {
					break
				}

				attempts += 1

				if idx.retry_after != 0 && attempts < idx.max_attempts {

					time_to_sleep := idx.retry_after * attempts

					slog.Error("Failed to walk URI, retry after delay", "attempts", attempts, "max_attempts", idx.max_attempts, "uri", uri, "error", walk_err, "seconds", time_to_sleep)

					time.Sleep(time.Duration(time_to_sleep) * time.Second)
				}
			}

			if walk_err != nil {
				slog.Error("Failed to walk URI, triggering error", "uri", uri, "error", walk_err)
				err_ch <- fmt.Errorf("Failed to walk '%s', %w", uri, walk_err)
			}

		}(uri)
	}

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return err
		default:
			// pass
		}
	}

	return nil
}

// IsIndexing() returns a boolean value indicating whether 'idx' is still processing documents.
func (idx *Iterator) IsIndexing() bool {

	if atomic.LoadInt64(&idx.count) > 0 {
		return true
	}

	return false
}

// increment() increments the count of documents being processed.
func (idx *Iterator) increment() {
	atomic.AddInt64(&idx.count, 1)
}

// decrement() decrements the count of documents being processed.
func (idx *Iterator) decrement() {
	atomic.AddInt64(&idx.count, -1)
}
