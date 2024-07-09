package main

// go run cmd/index-overture-venues/main.go /usr/local/data/overture/places-geojson/venues-0.95.geojsonl.bz2

import (
	"context"
	"flag"
	_ "fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/overture"
	"github.com/whosonfirst/go-dedupe/parser"
	"github.com/whosonfirst/go-overture/geojsonl"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	var location_database_uri string
	var location_parser_uri string
	var monitor_uri string
	var bucket_uri string
	var is_bzipped bool

	var start_after int

	flag.StringVar(&location_database_uri, "location-database-uri", "", "...")
	flag.StringVar(&location_parser_uri, "location-parser-uri", "overtureplaces://", "...")
	flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
	flag.StringVar(&bucket_uri, "bucket-uri", "file:///", "...")
	flag.BoolVar(&is_bzipped, "is-bzip2", true, "...")
	flag.IntVar(&start_after, "start-after", 0, "...")

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	db, err := location.NewDatabase(ctx, location_database_uri)

	if err != nil {
		log.Fatalf("Failed to create new location database, %v", err)
	}

	defer db.Close(ctx)

	prsr, err := parser.NewParser(ctx, location_parser_uri)

	if err != nil {
		log.Fatalf("Failed to create new location parser, %v", err)
	}

	source_bucket, err := bucket.OpenBucket(ctx, bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open source bucket, %v", err)
	}

	defer source_bucket.Close()

	monitor, err := timings.NewMonitor(ctx, monitor_uri)

	if err != nil {
		log.Fatalf("Failed to create monitor, %v", err)
	}

	monitor.Start(ctx, os.Stderr)
	defer monitor.Stop(ctx)

	max_workers := 20
	throttle := make(chan bool, max_workers)

	for i := 0; i < max_workers; i++ {
		throttle <- true
	}

	wg := new(sync.WaitGroup)

	walk_cb := func(ctx context.Context, path string, rec *walk.WalkRecord) error {

		if start_after > 0 && rec.LineNumber < start_after {
			monitor.Signal(ctx)
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
				// logger.Info("Done")
			}()

			loc, err := prsr.Parse(ctx, rec.Body)

			if dedupe.IsInvalidRecordError(err) {
				logger.Warn("Invalid record")
				return
			} else if err != nil {
				logger.Error("Failed to parse record", "error", err)
				return
				// return fmt.Errorf("Failed to parse body, %w", err)
			}

			logger = logger.With("id", loc.ID)
			logger = logger.With("location", loc)

			err = db.AddLocation(ctx, loc)

			if err != nil {
				logger.Error("Failed to add record", "error", err)
				return
				// return err
			}

			// logger.Info("OK", "geohash", loc.Geohash())
			monitor.Signal(ctx)

		}(path, rec)

		return nil
	}

	walk_opts := &geojsonl.WalkOptions{
		SourceBucket: source_bucket,
		Callback:     walk_cb,
		IsBzipped:    is_bzipped,
	}

	err = geojsonl.Walk(ctx, walk_opts, uris...)

	if err != nil {
		log.Fatalf("Failed to walk, %v", err)
	}

	wg.Wait()
}
