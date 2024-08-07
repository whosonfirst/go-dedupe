package dedupe

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/whosonfirst/go-dedupe/compare"
	"github.com/whosonfirst/go-dedupe/location"
)

// Compatator compares arbirtrary locations against a database of existing records.
type Comparator3 struct {
	source_database     location.Database
	target_database     location.Database
	vector_database_uri string
	workers             int
}

// ComparatorOptions is a struct containing configuration options used to create a new `Comparator` instance.
type Comparator3Options struct {
	SourceLocationDatabaseURI string
	TargetLocationDatabaseURI string
	VectorDatabaseURI         string
}

// NewComparator returns a new `Comparator` instance which wraps all the logic of comparing the embeddings
// for a given `Location` instance against a database of `Location` instances and emit matches as CSV rows.
func NewComparator3(ctx context.Context, opts *Comparator3Options) (*Comparator3, error) {

	source_db, err := location.NewDatabase(ctx, opts.SourceLocationDatabaseURI)

	if err != nil {
		return nil, fmt.Errorf("Failed to create location database, %w", err)
	}

	target_db, err := location.NewDatabase(ctx, opts.TargetLocationDatabaseURI)

	if err != nil {
		return nil, fmt.Errorf("Failed to create location database, %w", err)
	}

	workers := runtime.NumCPU() * 4
	slog.Info("WORKERS", "workers", workers)

	c := &Comparator3{
		source_database:     source_db,
		target_database:     target_db,
		vector_database_uri: opts.VectorDatabaseURI,
		workers:             workers,
	}

	return c, nil
}

// Compare compares 'body' against the database of existing records (contained by 'c'). Matches are written as CSV rows with the
// following keys: location (the location being compared), source (the matching source data that a location is compared against),
// similarity.
func (c *Comparator3) Compare(ctx context.Context, threshold float64) error {

	throttle := make(chan bool, c.workers)

	for i := 0; i < c.workers; i++ {
		throttle <- true
	}

	// For each geohash in the target database

	geohashes_cb := func(ctx context.Context, geohash string) error {

		<-throttle

		go func(geohash string) {

			defer func() {
				throttle <- true
			}()

			logger := slog.Default()
			logger = logger.With("geohash", geohash)

			logger.Debug("Process geohash")

			source_suffix := fmt.Sprintf("*-%s-source.jsonl", geohash)
			target_suffix := fmt.Sprintf("*-%s-target.jsonl", geohash)

			source_wr, err := os.CreateTemp("", source_suffix)

			if err != nil {
				logger.Error("Failed to create source writer", "error", err)
				return
			}

			source_path := source_wr.Name()
			defer os.Remove(source_path)

			target_wr, err := os.CreateTemp("", target_suffix)

			if err != nil {
				logger.Error("Failed to create target writer", "error", err)
				return
			}

			target_path := target_wr.Name()
			defer os.Remove(target_path)

			count_source := 0
			count_target := 0

			source_cb := func(ctx context.Context, loc *location.Location) error {

				enc_loc, err := json.Marshal(loc)

				if err != nil {
					logger.Warn("Failed to encode location, skipping", "id", loc.ID, "error", err)
					return nil
				}

				source_wr.Write(enc_loc)
				source_wr.Write([]byte("\n"))

				// slog.Debug("Write source", "id", loc.ID)
				count_source += 1
				return nil
			}

			err = c.source_database.GetWithGeohash(ctx, geohash, source_cb)

			if err != nil {
				return
			}

			target_cb := func(ctx context.Context, loc *location.Location) error {

				enc_loc, err := json.Marshal(loc)

				if err != nil {
					logger.Warn("Failed to encode location, skipping", "id", loc.ID, "error", err)
					return nil
				}

				target_wr.Write(enc_loc)
				target_wr.Write([]byte("\n"))

				count_target += 1
				return nil
			}

			err = c.target_database.GetWithGeohash(ctx, geohash, target_cb)

			if err != nil {
				return
			}

			source_root := filepath.Dir(source_path)
			source_fname := filepath.Base(source_path)

			target_root := filepath.Dir(target_path)
			target_fname := filepath.Base(target_path)

			source_bucket := fmt.Sprintf("file://%s", source_root)
			target_bucket := fmt.Sprintf("file://%s", target_root)

			logger.Info("Compare locations", "source", source_fname, "source count", count_source, "target", target_fname, "target count", count_target)

			compare_opts := &compare.CompareLocationsOptions{
				SourceBucketURI:   source_bucket,
				SourceLocations:   source_fname,
				TargetBucketURI:   target_bucket,
				TargetLocations:   target_fname,
				VectorDatabaseURI: c.vector_database_uri,
				Geohash:           geohash,
				Threshold:         threshold,
			}

			err = compare.CompareLocations(ctx, compare_opts)

			if err != nil {
				logger.Error("Failed to compare locations", "error", err)
			}

		}(geohash)

		return nil
	}

	slog.Debug("Get geohashes from target database")
	err := c.target_database.GetGeohashes(ctx, geohashes_cb)

	if err != nil {
		return err
	}

	return nil
}
