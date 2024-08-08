package main

/*

> go run cmd/index-locations/main.go -location-database-uri 'sql://sqlite3?dsn=/usr/local/data/overture/whosonfirst-locations.db&max-conns=1' -location-parser-uri 'whosonfirstvenues://' -iterator-uri 'whosonfirst://?iterator-uri=repo://' /usr/local/data/whosonfirst-data-venue-us-ca/

*/

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	// "sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	"github.com/whosonfirst/go-dedupe/iterator"
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/overture"
	"github.com/whosonfirst/go-dedupe/parser"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
)

func main() {

	var location_database_uri string
	var location_parser_uri string
	var iterator_uri string

	var monitor_uri string

	var verbose bool

	flag.StringVar(&location_database_uri, "location-database-uri", "", "...")
	flag.StringVar(&location_parser_uri, "location-parser-uri", "", "...")
	flag.StringVar(&iterator_uri, "iterator-uri", "", "...")

	flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
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
		log.Fatalf("Failed to create new location parser for '%s', %v", location_parser_uri, err)
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri)

	if err != nil {
		log.Fatalf("Failed to create iterator, %v", err)
	}

	defer func() {

		err := iter.Close(ctx)

		if err != nil {
			log.Fatalf("Failed to close iterator, %w", err)
		}
	}()

	monitor, err := timings.NewMonitor(ctx, monitor_uri)

	if err != nil {
		log.Fatalf("Failed to create monitor, %v", err)
	}

	monitor.Start(ctx, os.Stderr)
	defer monitor.Stop(ctx)

	iter_cb := func(ctx context.Context, body []byte) error {

		loc, err := prsr.Parse(ctx, body)

		if dedupe.IsInvalidRecordError(err) {
			slog.Warn("Invalid record")
			return nil
		} else if err != nil {
			slog.Error("Failed to parse record", "error", err)
			return fmt.Errorf("Failed to parse body, %w", err)
		}

		err = db.AddLocation(ctx, loc)

		if err != nil {
			slog.Error("Failed to add record", "error", err)
			return err
		}

		slog.Debug("Added location", "id", loc.ID, "location", loc.String())

		monitor.Signal(ctx)
		return nil
	}

	err = iter.IterateWithCallback(ctx, iter_cb, uris...)

	if err != nil {
		log.Fatalf("Failed to walk, %v", err)
	}
}
