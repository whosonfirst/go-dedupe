package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	_ "log/slog"
	"os"

	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	_ "github.com/whosonfirst/go-dedupe/overture"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
	_ "github.com/whosonfirst/go-writer-jsonl/v3"
	_ "gocloud.dev/blob/fileblob"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe/embeddings"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-overture/geojsonl"
	"github.com/whosonfirst/go-writer/v3"
)

type Row struct {
	ID         string            `json:"id"`
	Content    string            `json:"content"`
	Metadata   map[string]string `json:"metadata"`
	Embeddings []float64         `json:"float64"`
}

func main() {

	var parser_uri string
	var embedder_uri string
	var writer_uri string

	var monitor_uri string
	var bucket_uri string
	var is_bzipped bool

	flag.StringVar(&parser_uri, "parser-uri", "overtureplaces://", "...")
	flag.StringVar(&embedder_uri, "embedder-uri", "ollama://?model=mxbai-embed-large", "...")
	flag.StringVar(&writer_uri, "writer-uri", "jsonl://?writer=stdout://", "")

	flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
	flag.StringVar(&bucket_uri, "bucket-uri", "file:///", "...")
	flag.BoolVar(&is_bzipped, "is-bzip2", true, "...")

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	prsr, err := location.NewParser(ctx, parser_uri)

	if err != nil {
		log.Fatalf("Failed to create new parser, %v", err)
	}

	embdr, err := embeddings.NewEmbedder(ctx, embedder_uri)

	if err != nil {
		log.Fatalf("Failed to create new embedder, %v", err)
	}

	wr, err := writer.NewWriter(ctx, writer_uri)

	if err != nil {
		log.Fatalf("Failed to create writer, %v", err)
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

		loc, err := prsr.Parse(ctx, rec.Body)

		if err != nil {
			return fmt.Errorf("Failed to parse body, %w", err)
		}

		embeddings, err := embdr.Embeddings(ctx, loc.String())

		if err != nil {
			return fmt.Errorf("Failed to derive embeddings for %s, %w", path, err)
		}

		r := Row{
			ID:         loc.ID,
			Content:    loc.String(),
			Metadata:   loc.Metadata(),
			Embeddings: embeddings,
		}

		enc_r, err := json.Marshal(r)

		if err != nil {
			return fmt.Errorf("Failed to encode components, %w", err)
		}

		br := bytes.NewReader(enc_r)

		_, err = wr.Write(ctx, path, br)

		if err != nil {
			return fmt.Errorf("Failed to write encodings for %s, %w", path, err)
		}

		monitor.Signal(ctx)
		// slog.Info("OK", "path", path, "line", rec.LineNumber, "id", c.ID)
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
