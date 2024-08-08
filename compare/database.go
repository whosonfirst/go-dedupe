package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe/location"
)

type CompareLocationDatabasesOptions struct {
	SourceLocationDatabaseURI string
	TargetLocationDatabaseURI string
	VectorDatabaseURI         string
	MonitorURI                string
	Threshold                 float64
	Workers                   int
}

func CompareLocationDatabases(ctx context.Context, opts *CompareLocationDatabasesOptions) error {

	source_database, err := location.NewDatabase(ctx, opts.SourceLocationDatabaseURI)

	if err != nil {
		return fmt.Errorf("Failed to create location database, %w", err)
	}

	defer source_database.Close(ctx)

	target_database, err := location.NewDatabase(ctx, opts.TargetLocationDatabaseURI)

	if err != nil {
		return fmt.Errorf("Failed to create location database, %w", err)
	}

	defer target_database.Close(ctx)

	workers := opts.Workers

	if workers == 0 {
		workers = runtime.NumCPU()
	}

	slog.Debug("Set up workers", "count", workers)

	throttle := make(chan bool, workers)

	for i := 0; i < workers; i++ {
		throttle <- true
	}

	// For each geohash in the target database

	geohashes := make([]string, 0)

	geohashes_cb := func(ctx context.Context, geohash string) error {
		geohashes = append(geohashes, geohash)
		return nil
	}

	err = target_database.GetGeohashes(ctx, geohashes_cb)

	if err != nil {
		return err
	}

	count_geohashes := len(geohashes)
	slog.Info("Process geohashes", "count", count_geohashes)

	//

	monitor_uri := fmt.Sprintf("counter://PT60S?total=%d", count_geohashes)
	monitor, err := timings.NewMonitor(ctx, monitor_uri)

	if err != nil {
		return err
	}

	monitor.Start(ctx, os.Stderr)
	defer monitor.Stop(ctx)

	wg := new(sync.WaitGroup)

	for _, geohash := range geohashes {

		<-throttle

		wg.Add(1)

		go func(geohash string) {

			defer func() {
				monitor.Signal(ctx)
				throttle <- true
				wg.Done()
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

			source_opts := &WriteLocationsWithGeohashOptions{
				Database: source_database,
				Logger:   logger,
				Geohash:  geohash,
				Writer:   source_wr,
				Label:    "source",
			}

			count_source, err := WriteLocationsWithGeohash(ctx, source_opts)

			if err != nil {
				logger.Error("Failed to write source locations", "error", err)
				return
			}

			if count_source == 0 {
				logger.Debug("No source locations match geohash, skipping")
				return
			}

			target_opts := &WriteLocationsWithGeohashOptions{
				Database: target_database,
				Logger:   logger,
				Geohash:  geohash,
				Writer:   target_wr,
				Label:    "target",
			}

			count_target, err := WriteLocationsWithGeohash(ctx, target_opts)

			if err != nil {
				logger.Error("Failed to write target locations", "error", err)
				return
			}

			if count_target == 0 {
				logger.Debug("No target locations match geohash, skipping")
				return
			}

			source_root := filepath.Dir(source_path)
			source_fname := filepath.Base(source_path)

			target_root := filepath.Dir(target_path)
			target_fname := filepath.Base(target_path)

			// TBD – write source/target to buckets; for example if bucket config
			// in CompareLocationDatabasesOptions not empty then write files there
			// rather than defining explicit file:// URIs below.

			source_bucket := fmt.Sprintf("file://%s", source_root)
			target_bucket := fmt.Sprintf("file://%s", target_root)

			logger.Info("Compare locations", "source count", count_source, "target count", count_target)

			compare_opts := &CompareLocationsForGeohashOptions{
				SourceBucketURI:   source_bucket,
				SourceLocations:   source_fname,
				TargetBucketURI:   target_bucket,
				TargetLocations:   target_fname,
				VectorDatabaseURI: opts.VectorDatabaseURI,
				Geohash:           geohash,
				Threshold:         opts.Threshold,
			}

			err = CompareLocationsForGeohash(ctx, compare_opts)

			if err != nil {
				logger.Error("Failed to compare locations", "error", err)
			}

		}(geohash)

	}

	wg.Wait()
	return nil
}

type WriteLocationsWithGeohashOptions struct {
	Database location.Database
	Writer   io.Writer
	Logger   *slog.Logger
	Geohash  string
	Label    string
}

func WriteLocationsWithGeohash(ctx context.Context, opts *WriteLocationsWithGeohashOptions) (int, error) {

	count := 0

	cb_func := func(ctx context.Context, loc *location.Location) error {

		enc_loc, err := json.Marshal(loc)

		if err != nil {
			opts.Logger.Error("Failed to encode location", "label", opts.Label, "id", loc.ID, "error", err)
			return fmt.Errorf("Failed to encode %s location %s, %w", opts.Label, loc.ID, err)
		}

		opts.Writer.Write(enc_loc)
		opts.Writer.Write([]byte("\n"))

		count += 1
		return nil
	}

	opts.Logger.Debug("Get locations with geohash", "label", opts.Label)
	t1 := time.Now()

	err := opts.Database.GetWithGeohash(ctx, opts.Geohash, cb_func)

	if err != nil {
		return count, fmt.Errorf("Failed to get %s locations with geohash %s, %w", opts.Label, opts.Geohash, err)
	}

	opts.Logger.Debug("Got locations with geohash", "label", opts.Label, "count", count, "time", time.Since(t1))
	return count, nil
}
