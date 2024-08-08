package compare

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/sfomuseum/go-flags/flagset"
	wof_compare "github.com/whosonfirst/go-dedupe/compare"
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

	cmp_opts := &wof_compare.CompareLocationDatabasesOptions{
		SourceLocationDatabaseURI: source_location_database_uri,
		TargetLocationDatabaseURI: target_location_database_uri,
		VectorDatabaseURI:         vector_database_uri,
		Threshold:                 threshold,
	}

	err := wof_compare.CompareLocationDatabases(ctx, cmp_opts)

	if err != nil {
		return fmt.Errorf("Failed to compare location databases, %w", err)
	}

	return nil
}
