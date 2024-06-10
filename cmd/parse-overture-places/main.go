package main

// go run cmd/index-overture-venues/main.go /usr/local/data/overture/places-geojson/venues-0.95.geojsonl.bz2

import (
	"context"
	"flag"
	"fmt"
	"log"
	_ "log/slog"
	_ "os"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	_ "github.com/whosonfirst/go-dedupe/overture"
	"github.com/whosonfirst/go-dedupe/parser"
	"github.com/whosonfirst/go-overture/geojsonl"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	var parser_uri string
	var bucket_uri string
	var is_bzipped bool

	flag.StringVar(&parser_uri, "parser-uri", "overtureplaces://", "...")
	flag.StringVar(&bucket_uri, "bucket-uri", "file:///", "...")
	flag.BoolVar(&is_bzipped, "is-bzip2", true, "...")

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	prsr, err := parser.NewParser(ctx, parser_uri)

	if err != nil {
		log.Fatalf("Failed to create new parser, %v", err)
	}

	source_bucket, err := bucket.OpenBucket(ctx, bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open source bucket, %v", err)
	}

	defer source_bucket.Close()

	walk_cb := func(ctx context.Context, path string, rec *walk.WalkRecord) error {

		c, err := prsr.Parse(ctx, rec.Body)

		if err != nil {
			return fmt.Errorf("Failed to parse body, %w", err)
		}

		log.Println(c)
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
