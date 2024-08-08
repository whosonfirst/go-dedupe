package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
	_ "sync/atomic"
	"time"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-dedupe/vector"
	"github.com/whosonfirst/go-overture/geojsonl"
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
}

func CompareLocationsForGeohash(ctx context.Context, opts *CompareLocationsForGeohashOptions) error {

	logger := slog.Default()
	logger = logger.With("geohash", opts.Geohash)

	t0 := time.Now()

	defer func() {
		logger.Info("Time to compare locations", "time", time.Since(t0))
	}()

	logger.Debug("Compare locations", "source", opts.SourceLocations, "target", opts.TargetLocations)

	buf_writer := os.Stdout
	var csv_writer *csvdict.Writer

	mu := new(sync.RWMutex)

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

		// logger.Info("Add to vector database", "location", loc.String())
		err = vector_db.Add(ctx, loc)

		if err != nil {
			return fmt.Errorf("Failed to index location %s in vector db, %w", loc.ID, err)
		}

		count_sources += 1
		return nil
	}

	source_walk_opts := &geojsonl.WalkOptions{
		SourceBucket: source_bucket,
		Callback:     source_walk_cb,
	}

	t1 := time.Now()

	logger.Debug("Walk sources", "path", opts.SourceLocations)
	err = geojsonl.Walk(ctx, source_walk_opts, opts.SourceLocations)

	if err != nil {
		return fmt.Errorf("Failed to walk source locations, %w", err)
	}

	logger.Info("Time to index sources in vector db", "count", count_sources, "time", time.Since(t1))

	//

	target_walk_cb := func(ctx context.Context, path string, rec *walk.WalkRecord) error {

		var loc *location.Location

		err := json.Unmarshal(rec.Body, &loc)

		if err != nil {
			return fmt.Errorf("Failed to unmarshal record, %w", err)
		}

		geohash := opts.Geohash
		threshold := opts.Threshold

		logger.Debug("Compare location from target database", "geohash", geohash, "location", loc.String())

		// t1 := time.Now()

		results, err := vector_db.Query(ctx, loc)

		if err != nil {
			logger.Error("Failed to query", "geohash", geohash, "location", loc.String(), "error", err)
			return fmt.Errorf("Failed to query feature, %w", err)
		}

		// logger.Debug("Time to compare location from target database", "location", loc.String(), "time", time.Since(t1))

		for _, qr := range results {

			logger.Info("Possible", "geohash", geohash, "similarity", qr.Similarity, "wof", loc.String(), "ov", qr.Content)

			// Make this a toggle...
			if float64(qr.Similarity) > threshold {
				continue
			}

			if qr.ID == loc.ID {
				continue
			}

			logger.Info("Match", "geohash", geohash, "threshold", threshold, "similarity", qr.Similarity, "query", loc.String(), "candidate", qr.Content)

			row := map[string]string{
				"geohash":    geohash,
				"source_id":  qr.ID,
				"target_id":  loc.ID,
				"source":     qr.Content,
				"target":     loc.String(),
				"similarity": fmt.Sprintf("%02f", qr.Similarity),
			}

			mu.Lock()
			defer mu.Unlock()

			if csv_writer == nil {

				fieldnames := make([]string, 0)

				for k, _ := range row {
					fieldnames = append(fieldnames, k)
				}

				wr, err := csvdict.NewWriter(buf_writer, fieldnames)

				if err != nil {
					return fmt.Errorf("Failed to create CSV writer, %w", err)
				}

				err = wr.WriteHeader()

				if err != nil {
					return fmt.Errorf("Failed to write header for CSV writer, %w", err)
				}

				csv_writer = wr
			}

			err = csv_writer.WriteRow(row)

			if err != nil {
				return fmt.Errorf("Failed to write header for CSV writer, %w", err)
			}

			csv_writer.Flush()
			break
		}

		return nil
	}

	target_walk_opts := &geojsonl.WalkOptions{
		SourceBucket: target_bucket,
		Callback:     target_walk_cb,
	}

	t2 := time.Now()

	logger.Debug("Walk target", "path", opts.TargetLocations)
	err = geojsonl.Walk(ctx, target_walk_opts, opts.TargetLocations)

	if err != nil {
		return fmt.Errorf("Failed to walk target locations, %w", err)
	}

	logger.Info("Time to compare targets against source vector db", "time", time.Since(t2))

	// Write CSV buffer to opts.WriterBucket here...

	return nil
}
