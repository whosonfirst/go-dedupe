package process

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-dedupe"
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

	csv_dupes := fs.Args()

	r, err := reader.NewReader(ctx, reader_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new reader, %w", err)
	}

	wr, err := writer.NewWriter(ctx, writer_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new writer, %w", err)
	}

	id_prefix := fmt.Sprintf("%s:id=", dedupe.WHOSONFIRST_PREFIX)

	dupes := new(sync.Map)

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

			targets := []string{
				ns_source,
				ns_target,
			}

			sort.Strings(targets)
			key := strings.Join(targets, "-")

			_, seen := dupes.LoadOrStore(key, true)

			if seen {
				slog.Debug("Already processed, skipping", "key", key)
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

			logger := slog.Default()
			logger = logger.With("source", source_id)
			logger = logger.With("target", target_id)

			source_f, err := wof_reader.LoadBytes(ctx, r, source_id)

			if err != nil {
				logger.Error("Failed to load source record, skipping", "error", err)
				continue
			}

			target_f, err := wof_reader.LoadBytes(ctx, r, target_id)

			if err != nil {
				logger.Error("Failed to load target record, skipping", "error", err)
				continue
			}

			source_lastmod := properties.LastModified(source_f)
			target_lastmod := properties.LastModified(target_f)

			logger.Debug("Last mod", "source", source_lastmod, "target", target_lastmod)

			source_d := properties.Deprecated(source_f)
			target_d := properties.Deprecated(target_f)

			logger = logger.With("source deprecated", source_d)
			logger = logger.With("target deprecated", target_d)

			if source_d != "" && target_d != "" {
				continue
			}

			if source_d != "" {

				// START OF all of this could be put in an func(this, that) function

				source_superseded_by := properties.SupersededBy(source_f)

				if slices.Contains(source_superseded_by, target_id) {
					logger.Debug("Source is deprecated and already superseded by target, skipping")
					continue
				}

				logger.Info("Update superseded by for source", "source superseded_by", source_superseded_by)

				source_superseded_by = append(source_superseded_by, target_id)

				source_updates := map[string]interface{}{
					"properties.wof:superseded_by": source_superseded_by,
					"properties.mz:is_current":     0,
				}

				err := write_updates(ctx, wr, source_f, source_updates)

				if err != nil {
					slog.Error("Failed to write updates for source", "error", err)
				}

				logger.Info("Wrote superseded_by updates for source")

				// Check supersedes for target here

				target_supersedes := properties.Supersedes(target_f)

				if !slices.Contains(target_supersedes, source_id) {

					target_supersedes = append(target_supersedes, source_id)

					target_updates := map[string]interface{}{
						"properties.wof:supersedes": target_supersedes,
					}

					err := write_updates(ctx, wr, target_f, target_updates)

					if err != nil {
						slog.Error("Failed to write supersedes updates for target", "error", err)
					}

					logger.Info("Wrote supersedes updates for target")
				}

				// END OF all of this could be put in an func(this, that) function

				continue
			}

			if target_d != "" {

				// See notes about a func(this, that) function above

				target_superseded_by := properties.SupersededBy(target_f)

				if slices.Contains(target_superseded_by, source_id) {
					logger.Debug("Target is already superseded by source, skipping")
					continue
				}

				logger.Info("Update superseded by for target", "target superseded by", target_superseded_by)

				target_superseded_by = append(target_superseded_by, source_id)

				target_updates := map[string]interface{}{
					"properties.wof:superseded_by": target_superseded_by,
					"properties.mz:is_current":     0,
				}

				err := write_updates(ctx, wr, target_f, target_updates)

				if err != nil {
					slog.Error("Failed to write superseded_by updates for target", "error", err)
				}

				logger.Info("Wrote superseded_by updates for target")

				source_supersedes := properties.Supersedes(source_f)

				if !slices.Contains(source_supersedes, target_id) {

					source_supersedes = append(source_supersedes, target_id)

					source_updates := map[string]interface{}{
						"properties.wof:supersedes": source_supersedes,
					}

					err := write_updates(ctx, wr, source_f, source_updates)

					if err != nil {
						slog.Error("Failed to write supersedes updates for source", "error", err)
					}

					logger.Info("Wrote supersedes updates for source")
				}

				continue
			}

			// logger.Info("PROCESS")
		}
	}

	return nil
}

func write_updates(ctx context.Context, wr writer.Writer, body []byte, updates map[string]interface{}) error {

	has_changes, new_body, err := export.AssignPropertiesIfChanged(ctx, body, updates)

	if err != nil {
		return fmt.Errorf("Failed to assign properties, %w", err)
	}

	if has_changes {

		_, err := wof_writer.WriteBytes(ctx, wr, new_body)

		if err != nil {
			return fmt.Errorf("Failed to write updates, %w", err)
		}
	}

	return nil
}
