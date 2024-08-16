package compare

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/url"

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

	// START of flag replacement hoohah because URL query escaping makes your eyes sad

	if vector_database_dsn != "" || vector_database_embedder_uri != "" || vector_database_model != "" {

		slog.Debug("Process replacement flags", "vector-database-dsn", vector_database_dsn, "vector-database-embedder-uri", vector_database_embedder_uri, "vector-database-model", vector_database_model)

		u, err := url.Parse(vector_database_uri)

		if err != nil {
			return fmt.Errorf("Failed to parse -vector-database-uri flag, %w", err)
		}

		q := u.Query()

		if q.Get("model") == "{vector-database-model}" {

			if vector_database_model == "" {
				return fmt.Errorf("Missing or empty -vector-database-model flag")
			}

			slog.Debug("Process {vector-database-model} query replacement", "model", vector_database_model)

			q.Del("model")
			q.Set("model", vector_database_model)
		}

		if q.Get("dsn") == "{vector-database-dsn}" {

			if vector_database_dsn == "" {
				return fmt.Errorf("Missing or empty -vector-database-dsn flag")
			}

			slog.Debug("Process {vector-database-dsn} query replacement", "dsn", vector_database_dsn)

			q.Del("dsn")
			q.Set("dsn", vector_database_dsn)
		}

		if q.Get("embedder-uri") == "{vector-database-embedder-uri}" {

			if vector_database_embedder_uri == "" {
				return fmt.Errorf("Missing or empty -vector-database-embedder-uri flag")
			}

			if vector_database_model != "" {

				eu, err := url.Parse(vector_database_embedder_uri)

				if err != nil {
					return fmt.Errorf("Failed to parse -vector-database-embedder-uri flag, %w", err)
				}

				eq := eu.Query()

				if eq.Get("model") == "{vector-database-model}" {

					slog.Debug("Process {vector-database-model} query replacement in embedder URI", "model", vector_database_model)

					eq.Del("model")
					eq.Set("model", vector_database_model)

					eu.RawQuery = eq.Encode()

					slog.Debug("Rewriting vector database embedder URI", "uri", eu.String())
					vector_database_embedder_uri = eu.String()
				}
			}

			slog.Debug("Process {vector-database-embedder-uri} query replacement", "embedder uri", vector_database_embedder_uri)

			q.Del("embedder-uri")
			q.Set("embedder-uri", vector_database_embedder_uri)
		}

		u.RawQuery = q.Encode()

		slog.Debug("Rewriting vector database URI", "uri", u.String())
		vector_database_uri = u.String()
	}

	// END of flag replacement hoohah because URL query escaping makes your eyes sad

	cmp_opts := &wof_compare.CompareLocationDatabasesOptions{
		SourceLocationDatabaseURI: source_location_database_uri,
		TargetLocationDatabaseURI: target_location_database_uri,
		VectorDatabaseURI:         vector_database_uri,
		Workers:                   workers,
		Threshold:                 threshold,
	}

	err := wof_compare.CompareLocationDatabases(ctx, cmp_opts)

	if err != nil {
		return fmt.Errorf("Failed to compare location databases, %w", err)
	}

	return nil
}
