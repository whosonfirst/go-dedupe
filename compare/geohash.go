package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	// "os"
	"strings"
	// "sync"
	// "sync/atomic"
	"time"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-dedupe/vector"
	// "github.com/whosonfirst/go-overture/geojsonl"
)

type CompareLocationsForGeohashOptions struct {
	SourceBucketURI   string
	SourceLocations   string
	TargetBucketURI   string
	TargetLocations   string
	WriterBucketURI   string
	WriterPrefix      string
	VectorDatabaseURI string
	Geohash           string
	Threshold         float64
	RowChannel        chan (map[string]string)
}

func CompareLocationsForGeohash(ctx context.Context, opts *CompareLocationsForGeohashOptions) error {

	logger := slog.Default()
	logger = logger.With("geohash", opts.Geohash)

	t0 := time.Now()

	defer func() {
		logger.Info("Time to compare locations", "time", time.Since(t0))
	}()

	logger.Debug("Compare locations", "source", opts.SourceLocations, "target", opts.TargetLocations)

	// buf_writer := os.Stdout
	// var csv_writer *csvdict.Writer
	// mu := new(sync.RWMutex)

	// Set up buckets

	source_bucket, err := bucket.OpenBucket(ctx, opts.SourceBucketURI)

	if err != nil {
		return err
	}

	defer source_bucket.Close()

	target_bucket, err := bucket.OpenBucket(ctx, opts.TargetBucketURI)

	if err != nil {
		return err
	}

	defer target_bucket.Close()

	// Create the vector database

	db_uri, _ := url.QueryUnescape(opts.VectorDatabaseURI)
	db_uri = strings.Replace(db_uri, "{geohash}", opts.Geohash, 1)

	vector_db, err := vector.NewDatabase(ctx, db_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new database, %w", err)
	}

	defer vector_db.Close(ctx)

	// Populate the vector database

	count_sources := 0

	source_walk_cb := func(ctx context.Context, path string, rec *walk.WalkRecord) error {

		var loc *location.Location

		err := json.Unmarshal(rec.Body, &loc)

		if err != nil {
			return fmt.Errorf("Failed to unmarshal record, %w", err)
		}

		logger.Debug("Add to vector database", "location", loc.String())
		err = vector_db.Add(ctx, loc)

		if err != nil {
			return fmt.Errorf("Failed to index location %s in vector db, %w", loc.ID, err)
		}

		count_sources += 1
		return nil
	}

	source_r, err := source_bucket.NewReader(ctx, opts.SourceLocations, nil)

	if err != nil {
		return err
	}

	defer source_r.Close()

	t1 := time.Now()

	// logger.Info("Walk sources", "path", opts.SourceLocations)
	err = walk_reader(ctx, source_r, source_walk_cb)

	if err != nil {
		return fmt.Errorf("Failed to walk source locations, %w", err)
	}

	logger.Info("Time to index sources in vector db", "count", count_sources, "time", time.Since(t1))

	target_walk_cb := func(ctx context.Context, path string, rec *walk.WalkRecord) error {

		var loc *location.Location

		err := json.Unmarshal(rec.Body, &loc)

		if err != nil {
			return fmt.Errorf("Failed to unmarshal record, %w", err)
		}

		geohash := opts.Geohash
		threshold := opts.Threshold

		logger.Debug("Compare location from target database", "location", loc.String())

		// t1 := time.Now()

		results, err := vector_db.Query(ctx, loc)

		if err != nil {
			logger.Error("Failed to query", "location", loc.String(), "error", err)
			return fmt.Errorf("Failed to query feature, %w", err)
		}

		// logger.Debug("Time to compare location from target database", "location", loc.String(), "time", time.Since(t1))

		for _, qr := range results {

			if qr.ID == loc.ID {
				continue
			}

			logger.Debug("Possible", "similarity", qr.Similarity, "wof", loc.String(), "ov", qr.Content)

			ok, err := vector_db.MeetsThreshold(ctx, qr, threshold)

			if err != nil {
				logger.Error("Failed to determine if query result meets threshold", "id", qr.ID, "error", err)
				continue
			}

			if !ok {
				continue
			}

			logger.Info("Match", "threshold", threshold, "similarity", qr.Similarity, "query", loc.String(), "candidate", qr.Content)

			row := map[string]string{
				"geohash":    geohash,
				"source_id":  qr.ID,
				"target_id":  loc.ID,
				"source":     qr.Content,
				"target":     loc.String(),
				"similarity": fmt.Sprintf("%02f", qr.Similarity),
			}

			opts.RowChannel <- row
			break
		}

		return nil
	}

	target_r, err := target_bucket.NewReader(ctx, opts.TargetLocations, nil)

	if err != nil {
		return err
	}

	defer target_r.Close()

	t2 := time.Now()

	logger.Info("Walk targets", "path", opts.TargetLocations)
	err = walk_reader(ctx, target_r, target_walk_cb)

	if err != nil {
		return fmt.Errorf("Failed to walk target locations, %w", err)
	}

	logger.Info("Time to index targets in vector db", "time", time.Since(t2))

	return nil
}

func walk_reader(ctx context.Context, r io.Reader, cb func(ctx context.Context, path string, rec *walk.WalkRecord) error) error {

	// walk_ctx, cancel := context.WithCancel(ctx)
	// defer cancel()

	var walk_err error

	record_ch := make(chan *walk.WalkRecord)
	error_ch := make(chan *walk.WalkError)
	done_ch := make(chan bool)

	go func() {

		for {
			select {
			case <-ctx.Done():
				done_ch <- true
				return
			case err := <-error_ch:
				slog.Error("Walk error", "error", err)
				walk_err = err
				done_ch <- true
			case r := <-record_ch:

				err := cb(ctx, r.Path, r)

				r.CompletedChannel <- true

				if err != nil {
					error_ch <- &walk.WalkError{
						Path:       r.Path,
						LineNumber: r.LineNumber,
						Err:        fmt.Errorf("Failed to index feature, %w", err),
					}
				}
			}
		}
	}()

	walk_opts := &walk.WalkOptions{
		RecordChannel: record_ch,
		ErrorChannel:  error_ch,
		DoneChannel:   done_ch,
		// This is necessary in order to force the jsonl/walk code to block
		// long enough for the callbacks above to execute. For very small
		// comparison tasks (like 1-5 records) it can happen that the walk
		// process completes as soon as all the r := <-record_ch events
		// have been received and dispatching walk_opts.DoneChannel before
		// any work happens. I was under the impression that r := <-record_ch
		// blocks but apparently not.
		SendCompletedChannel: true,
		Workers:              10,
	}

	go walk.WalkReader(ctx, walk_opts, r)

	<-walk_opts.DoneChannel

	if walk_err != nil && !walk.IsEOFError(walk_err) {
		return fmt.Errorf("Failed to walk document, %v", walk_err)
	}

	return nil
}
