package main

// go run cmd/index-overture-locations/main.go -location-database-uri 'sql://sqlite3?dsn=/usr/local/data/overture/overture-locations.db' /usr/local/data/overture/places-geojson/venues-0.95.geojsonl.bz2

import (
	"context"
	"flag"
	_ "fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-dedupe/parser"
)

func main() {

	var location_database_uri string
	var location_parser_uri string
	var monitor_uri string

	// var start_after int
	var verbose bool

	flag.StringVar(&location_database_uri, "location-database-uri", "", "...")
	flag.StringVar(&location_parser_uri, "location-parser-uri", "alltheplaces://", "...")
	flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
	// flag.IntVar(&start_after, "start-after", 0, "...")
	flag.BoolVar(&verbose, "verbose", false, "...")

	flag.Parse()

	uris := flag.Args()

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

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

	for _, path := range uris {

		r, err := os.Open(path)

		if err != nil {
			log.Fatalf("Failed to open %s for reading, %v", path, err)
		}

		defer r.Close()

		body, err := io.ReadAll(r)

		if err != nil {
			log.Fatalf("Failed to read %s, %v", path, err)
		}

		slog.Info("Unmarshal", "path", path)

		fc, err := geojson.UnmarshalFeatureCollection(body)

		if err != nil {
			slog.Warn("Failed to unmarshal feature collection", "path", path, "error", err)
			continue
		}

		for _, f := range fc.Features {

			<-throttle

			wg.Add(1)

			body, err := f.MarshalJSON()

			if err != nil {
				slog.Error("Failed to marshal record", "error", err)
				continue
			}

			loc, err := prsr.Parse(ctx, body)

			if dedupe.IsInvalidRecordError(err) {
				slog.Warn("Invalid record")
				continue
			} else if err != nil {
				slog.Error("Failed to parse record", "error", err)
				continue
				// return fmt.Errorf("Failed to parse body, %w", err)
			}

			err = db.AddLocation(ctx, loc)

			if err != nil {
				slog.Error("Failed to add record", "error", err)
				continue
				// return err
			}

			monitor.Signal(ctx)
		}
	}

	wg.Wait()
}
