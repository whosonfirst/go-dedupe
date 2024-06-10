package main

// go run cmd/index-overture-venues/main.go /usr/local/data/overture/places-geojson/venues-0.95.geojsonl.bz2

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/database"
	_ "github.com/whosonfirst/go-dedupe/overture"
	"github.com/whosonfirst/go-dedupe/parser"
	"github.com/whosonfirst/go-overture/geojsonl"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	var database_uri string
	var parser_uri string
	var monitor_uri string
	var bucket_uri string
	var is_bzipped bool

	flag.StringVar(&database_uri, "database-uri", "chromem://venues/usr/local/data/venues.db?model=mxbai-embed-large", "...")
	flag.StringVar(&parser_uri, "parser-uri", "overtureplaces://", "...")
	flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
	flag.StringVar(&bucket_uri, "bucket-uri", "file:///", "...")
	flag.BoolVar(&is_bzipped, "is-bzip2", true, "...")

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	prsr, err := parser.NewParser(ctx, parser_uri)

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

	walk_cb := func(ctx context.Context, path string, rec *walk.WalkRecord) error {

		logger := slog.Default()
		logger = logger.With("path", path)
		logger = logger.With("line number", rec.LineNumber)

		loc, err := prsr.Parse(ctx, rec.Body)

		if dedupe.IsInvalidRecordError(err) {
			logger.Warn("Invalid record")
			return nil
		} else if err != nil {
			return fmt.Errorf("Failed to parse body, %w", err)
		}

		logger = logger.With("location", loc)

		err = db.Add(ctx, loc.ID, loc.Content(), loc.Metadata())

		if err != nil {
			logger.Error("Failed to add record", "error", err)
			return err
		}

		monitor.Signal(ctx)
		logger.Info("OK")
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
}
