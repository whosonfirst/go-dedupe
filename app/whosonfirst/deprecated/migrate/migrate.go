package migrate

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	_ "github.com/whosonfirst/go-whosonfirst-uri"
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

	target_repo_name := filepath.Base(target_repo)

	target_writer_uri := fmt.Sprintf("repo://%s", target_repo)

	target_wr, err := writer.NewWriter(ctx, target_writer_uri)

	if err != nil {
		return fmt.Errorf("Failed to create target writer, %w", err)
	}

	source_iterator_uri := fmt.Sprintf("repo://%s?include=properties.edtf:deprecated=.*", source_repo)

	source_iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		slog.Info("Process record", "path", path)

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		updates := map[string]interface{}{
			"properties.wof:repo": target_repo_name,
		}

		new_body, err := export.AssignProperties(ctx, body, updates)

		if err != nil {
			return fmt.Errorf("Failed to assign new properties to %s, %w", path, err)
		}

		_, err = wof_writer.WriteBytes(ctx, target_wr, new_body)

		if err != nil {
			slog.Warn("Failed to write updates", "path", path, "error", err)
			return nil
			// return fmt.Errorf("Failed to assign write updates for %s, %w", path, err)
		}

		err = os.Remove(path)

		if err != nil {
			return fmt.Errorf("Failed to remove record", "path", path)
		}

		return nil
	}

	source_iter, err := iterator.NewIterator(ctx, source_iterator_uri, source_iter_cb)

	if err != nil {
		return fmt.Errorf("Failed to create source iterator, %w", err)
	}

	err = source_iter.IterateURIs(ctx, source_repo)

	if err != nil {
		return fmt.Errorf("Failed to iterate %s, %w", source_repo, err)
	}

	return nil
}
