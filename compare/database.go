package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe/location"
)

type CompareLocationDatabasesOptions struct {
	SourceLocationDatabaseURI string
	TargetLocationDatabaseURI string
	VectorDatabaseURI         string
	MonitorURI                string
	Threshold                 float64
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

	workers := runtime.NumCPU() * 4
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

			err = source_database.GetWithGeohash(ctx, geohash, source_cb)

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

			err = target_database.GetWithGeohash(ctx, geohash, target_cb)

			if err != nil {
				return
			}

			source_root := filepath.Dir(source_path)
			source_fname := filepath.Base(source_path)

			target_root := filepath.Dir(target_path)
			target_fname := filepath.Base(target_path)

			source_bucket := fmt.Sprintf("file://%s", source_root)
			target_bucket := fmt.Sprintf("file://%s", target_root)

			// logger.Info("Compare locations", "source", source_path, "source count", count_source, "target", target_path, "target count", count_target)
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
