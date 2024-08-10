package process

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-reader"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	// wof_writer "github.com/whosonfirst/go-whosonfirst-writer/v3"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
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

	csv_dupes := fs.Args()

	r, err := reader.NewReader(ctx, reader_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new reader, %w", err)
	}

	wr, err := writer.NewWriter(ctx, writer_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new writer, %w", err)
	}

	slog.Debug("debug", "wr", wr)

	id_prefix := fmt.Sprintf("%s:id=", dedupe.WHOSONFIRST_PREFIX)

	for _, path := range csv_dupes {

		csv_r, err := csvdict.NewReaderFromPath(path)

		if err != nil {
			return fmt.Errorf("Failed to open %s as CSV file, %w", path, err)
		}

		for {
			row, err := csv_r.Read()

			if err == io.EOF {
				break
			}

			if err != nil {
				return fmt.Errorf("CSV reader reported an error parsing %s, %w", path, err)
			}

			ns_source, ok := row["source_id"]

			if !ok {
				slog.Warn("Row is missing source ID, skipping", "row", row)
				continue
			}

			ns_target, ok := row["target_id"]

			if !ok {
				slog.Warn("Row is missing target ID, skipping", "row", row)
				continue
			}

			str_source := strings.Replace(ns_source, id_prefix, "", 1)
			str_target := strings.Replace(ns_target, id_prefix, "", 1)

			source_id, err := strconv.ParseInt(str_source, 10, 64)

			if err != nil {
				slog.Error("Failed to parse source ID, skipping", "id", str_source, "id (ns)", ns_source, "error", err)
				continue
			}

			target_id, err := strconv.ParseInt(str_target, 10, 64)

			if err != nil {
				slog.Error("Failed to parse target ID, skipping", "id", str_target, "id (ns)", ns_target, "error", err)
				continue
			}

			source_f, err := wof_reader.LoadBytes(ctx, r, source_id)

			if err != nil {
				slog.Error("Failed to load source record, skipping", "id", source_id, "error", err)
				continue
			}

			target_f, err := wof_reader.LoadBytes(ctx, r, target_id)

			if err != nil {
				slog.Error("Failed to load target record, skipping", "id", target_id, "error", err)
				continue
			}

			source_lastmod := properties.LastModified(source_f)
			target_lastmod := properties.LastModified(target_f)

			slog.Info("Last mod", "source", source_lastmod, "target", target_lastmod)
		}
	}

	return nil
}
