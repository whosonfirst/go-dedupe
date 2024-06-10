package main

/*

[asc][asc@SD-931-4][11:05:22] /usr/local/whosonfirst/go-whosonfirst-dedupe                                                                                                                     > go run cmd/compare-alltheplaces/main.go /usr/local/data/alltheplaces/dunkin_us.geojson
2024/06/09 11:05:23 INFO Create database
2024/06/09 11:17:10 Failed to create new database, Failed to create database, couldn't read document: couldn't open file: open /usr/local/data/venues.db/bc318ecb/173de102.gob: too many open files
exit status 1

*/

import (
	"context"
	"flag"
	// "fmt"
	"io"
	"log"
	"log/slog"
	"os"

	// "github.com/aaronland/go-jsonl/walk"
	// "github.com/aaronland/gocloud-blob/bucket"
	// "github.com/sfomuseum/go-timings"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	"github.com/whosonfirst/go-dedupe/database"
	_ "github.com/whosonfirst/go-dedupe/overture"
	"github.com/whosonfirst/go-dedupe/parser"
	// "github.com/whosonfirst/go-overture/geojsonl"
	"github.com/paulmach/orb/geojson"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	var database_uri string
	var parser_uri string
	// var monitor_uri string
	// var bucket_uri string
	// var is_bzipped bool

	flag.StringVar(&database_uri, "database-uri", "chromem://venues/usr/local/data/venues.db?model=mxbai-embed-large", "...")
	flag.StringVar(&parser_uri, "parser-uri", "alltheplaces://", "...")
	// flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
	// flag.StringVar(&bucket_uri, "bucket-uri", "file:///", "...")
	// flag.BoolVar(&is_bzipped, "is-bzip2", true, "...")

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	slog.Info("Create database")
	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	slog.Info("Create parser")
	prsr, err := parser.NewParser(ctx, parser_uri)

	if err != nil {
		log.Fatalf("Failed to create new parser, %v", err)
	}

	/*
		source_bucket, err := bucket.OpenBucket(ctx, bucket_uri)

		if err != nil {
			log.Fatalf("Failed to open source bucket, %v", err)
		}
	*/

	for _, path := range uris {

		r, err := os.Open(path)

		if err != nil {
			log.Fatalf("Failed to open %s for reading, %v", path, err)
		}

		defer r.Close()

		body, err := io.ReadAll(r)

		if err != nil {
			log.Fatalf("Failed to read %s, %v", path, err)
		}

		slog.Info("Unmarshal", "path", path)

		fc, err := geojson.UnmarshalFeatureCollection(body)

		if err != nil {
			log.Fatalf("Failed to unmarshal %s, %v", path, err)
		}

		for idx, f := range fc.Features {

			f_body, err := f.MarshalJSON()

			if err != nil {
				log.Fatalf("Failed to marshal feature at offset %d for %s, %v", idx, path, err)
			}

			c, err := prsr.Parse(ctx, f_body)

			if err != nil {
				log.Fatalf("Failed to parse feature at offset %d for %s, %v", idx, path, err)
			}

			results, err := db.Query(ctx, c.Content())

			if err != nil {
				log.Fatalf("Failed to query for feature at offset %d for %s: %s, %v", idx, path, c.Content, err)
			}

			slog.Info("results", "path", path, "offset", idx, "query", c.Content, "results", len(results))

			for _, qr := range results {

				log.Printf("[%s][%s] %0.6f %s %s\n", c.ID, c.Content, qr.Similarity, qr.ID, qr.Content)
			}
		}

	}

}
