package assign

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer/v3"
	"github.com/whosonfirst/go-writer/v3"
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

	r, err := reader.NewReader(ctx, reader_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new reader, %w", err)
	}

	wr, err := writer.NewWriter(ctx, writer_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new writer, %w", err)
	}

	concordance_key := fmt.Sprintf("%s:%s", concordance_namespace, concordance_predicate)

	concordances := fs.Args()

	for _, path := range concordances {

		csv_r, err := csvdict.NewReaderFromPath(path)

		if err != nil {
			return fmt.Errorf("Failed to create new CSV reader for %s, %w", path, err)
		}

		for {
			row, err := csv_r.Read()

			if err == io.EOF {
				break
			}

			logger := slog.Default()
			logger = logger.With("source", row["source_id"])
			logger = logger.With("target", row["target_id"])
			logger = logger.With("wof label", wof_label)

			var str_wof_id string
			var other_id string
			var other_label string

			switch wof_label {
			case "target":
				str_wof_id = row["target_id"]
				other_id = row["source_id"]
				other_label = row["source"]

			default:
				str_wof_id = row["source_id"]
				other_id = row["target_id"]
				other_label = row["target"]
			}

			// Please make this less bad
			other_parts := strings.Split(other_id, "=")
			other_id = other_parts[1]

			// Please make this less bad
			str_wof_id = strings.Replace(str_wof_id, "wof:id=", "", 1)

			wof_id, err := strconv.ParseInt(str_wof_id, 10, 64)

			if err != nil {
				logger.Warn("Failed to parse WOF ID, skipping", "error", err)
				continue
			}

			logger = logger.With("wof id", wof_id)
			logger = logger.With("other id", other_id)

			body, err := wof_reader.LoadBytes(ctx, r, wof_id)

			if err != nil {
				logger.Error("Failed to load body for WOF record", "error", err)
				continue
			}

			concordances := properties.Concordances(body)
			concordances[concordance_key] = other_id

			updates := map[string]interface{}{
				"properties.wof:concordances":                                  concordances,
				fmt.Sprintf("properties.%s:label", concordance_namespace):      other_label,
				fmt.Sprintf("properties.%s:similarity", concordance_namespace): row["similarity"],
			}

			has_changes, new_body, err := export.AssignPropertiesIfChanged(ctx, body, updates)

			if err != nil {
				logger.Error("Failed to assign new properties, skipping", "error", err)
				continue
			}

			if !has_changes {
				logger.Debug("Record has no changes, skipping")
				continue
			}

			_, err = wof_writer.WriteBytes(ctx, wr, new_body)

			if err != nil {
				logger.Error("Failed to write changes", "error", err)
			}

			logger.Debug("Updates concordances")
		}
	}

	return nil
}
