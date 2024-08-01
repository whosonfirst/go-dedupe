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
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe"
	_ "github.com/whosonfirst/go-dedupe/overture"
	"github.com/whosonfirst/go-dedupe/parser"
	"github.com/whosonfirst/go-dedupe/vector"
	"github.com/whosonfirst/go-overture/geojsonl"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	var vector_database_uri string
	var location_parser_uri string
	var monitor_uri string
	var bucket_uri string
	var is_bzipped bool
	var verbose bool
	
	var start_after int

	//flag.StringVar(&vector_database_uri, "vector-database-uri", "chromem://venues/usr/local/data/venues.db?model=mxbai-embed-large", "...")
	flag.StringVar(&vector_database_uri, "vector-database-uri", "sqlite://?model=mxbai-embed-large&dsn=/usr/local/data/overture/vectors.db&embedder-uri=ollama%3A%2F%2F%3Fmodel%3Dmxbai-embed-large&max-distance=4&max-results=10&dimensions=1024", "...")	
	flag.StringVar(&location_parser_uri, "location-parser-uri", "overtureplaces://", "...")
	flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
	flag.StringVar(&bucket_uri, "bucket-uri", "file:///", "...")
	flag.BoolVar(&is_bzipped, "is-bzip2", true, "...")
	flag.IntVar(&start_after, "start-after", 0, "...")
	flag.BoolVar(&verbose, "verbose", false, "...")
	
	flag.Parse()

	uris := flag.Args()

	if verbose {
                slog.SetLogLoggerLevel(slog.LevelDebug)
                slog.Debug("Verbose logging enabled")
	}
	
	ctx := context.Background()

	db, err := vector.NewDatabase(ctx, vector_database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	defer db.Close(ctx)

	prsr, err := parser.NewParser(ctx, location_parser_uri)

	if err != nil {
		log.Fatalf("Failed to create new parser, %v", err)
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

			err = db.Add(ctx, loc)

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

	err = db.Flush(ctx)

	if err != nil {
		log.Fatalf("Failed to flush database, %v", err)
	}
}
