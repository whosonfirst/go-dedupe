package main

/*

[asc][asc@SD-931-4][11:05:22] /usr/local/whosonfirst/go-whosonfirst-dedupe                                                                                                                     > go run cmd/compare-alltheplaces/main.go /usr/local/data/alltheplaces/dunkin_us.geojson

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

	var threshold float64

	flag.StringVar(&database_uri, "database-uri", "opensearch://?dsn=https%3A%2F%2Flocalhost%3A9200%2Fdedupe%3Fusername%3Dadmin%26password%3DKJHFGDFJGSJfsdkjfhsdoifruwo45978h52dcn%26insecure%3Dtrue%26require-tls%3Dtrue&model=a8-aBJABf__qJekL_zJC&bulk-index=false", "...")

	//flag.StringVar(&database_uri, "database-uri", "chromem://venues/usr/local/data/venues.db?model=mxbai-embed-large", "...")
	flag.StringVar(&parser_uri, "parser-uri", "alltheplaces://", "...")
	// flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
	// flag.StringVar(&bucket_uri, "bucket-uri", "file:///", "...")
	// flag.BoolVar(&is_bzipped, "is-bzip2", true, "...")

	flag.Float64Var(&threshold, "threshold", 0.95, "...")
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

	total_matches := 0
	total_features := 0

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
			slog.Warn("Failed to unmarshal feature collection", "path", path, "error", err)
			continue
		}

		features := 0
		matches := 0

		for idx, f := range fc.Features {

			features += 1
			total_features += 1

			logger := slog.Default()
			logger = logger.With("path", path)
			logger = logger.With("offset", idx)

			f_body, err := f.MarshalJSON()

			if err != nil {
				logger.Warn("Failed to marshal feature", "error", err)
				continue
			}

			c, err := prsr.Parse(ctx, f_body)

			if err != nil {
				logger.Warn("Failed to parse feature", "error", err)
				continue
			}

			results, err := db.Query(ctx, c.Content(), c.Metadata())

			if err != nil {
				logger.Warn("Failed to query feature", "error", err)
				continue
			}

			for _, qr := range results {

				if float64(qr.Similarity) >= threshold {
					logger.Info("Match", "similarity", qr.Similarity, "atp", c.Content(), "ov", qr.Content)
					matches += 1
					total_matches += 1
					break
				}
			}
		}

		slog.Info("Matches", "path", path, "features", features, "matches", matches, "total features", total_features, "total matches", total_matches)

	}

}
