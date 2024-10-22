package index

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/iterator"
	"github.com/whosonfirst/go-dedupe/location"
)

func Run(ctx context.Context) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	uris := fs.Args()

	db, err := location.NewDatabase(ctx, location_database_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new location database, %v", err)
	}

	defer db.Close(ctx)

	prsr, err := location.NewParser(ctx, location_parser_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new location parser for '%s', %v", location_parser_uri, err)
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri)

	if err != nil {
		return fmt.Errorf("Failed to create iterator, %v", err)
	}

	defer func() {

		err := iter.Close(ctx)

		if err != nil {
			slog.Error("Failed to close iterator", "error", err)
		}
	}()

	monitor, err := timings.NewMonitor(ctx, monitor_uri)

	if err != nil {
		return fmt.Errorf("Failed to create monitor, %v", err)
	}

	monitor.Start(ctx, os.Stderr)
	defer monitor.Stop(ctx)

	iter_cb := func(ctx context.Context, body []byte) error {

		loc, err := prsr.Parse(ctx, body)

		if dedupe.IsInvalidRecordError(err) {
			// slog.Warn("Invalid record")
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
		return fmt.Errorf("Failed to walk, %v", err)
	}

	return nil
}
