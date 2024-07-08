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
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/whosonfirst/go-dedupe"
	_ "github.com/whosonfirst/go-dedupe/overture"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

func main() {

	var embeddings_database_uri string
	var location_database_uri string
	var location_parser_uri string
	var iterator_uri string
	var threshold float64

	flag.StringVar(&embeddings_database_uri, "embeddings-database-uri", "chromem://{geohash}?model=mxbai-embed-large", "...")

	flag.StringVar(&location_database_uri, "location-database-uri", "", "...")
	flag.StringVar(&location_parser_uri, "parser-uri", "alltheplaces://", "...")

	flag.StringVar(&iterator_uri, "iterator-uri", "repo://", "...")

	flag.Float64Var(&threshold, "threshold", 0.75, "...")
	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	cmp_opts := &dedupe.ComparatorOptions{
		LocationDatabaseURI:   location_database_uri,
		LocationParserURI:     location_parser_uri,
		EmbeddingsDatabaseURI: embeddings_database_uri,
		Writer:                os.Stdout,
	}

	cmp, err := dedupe.NewComparator(ctx, cmp_opts)

	if err != nil {
		log.Fatalf("Failed to create new comparator, %v", err)
	}

	defer cmp.Flush()

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %v", path, err)
		}

		is_match, err := cmp.Compare(ctx, body, threshold)

		if err != nil {
			slog.Warn("Failed to compare feature", "path", path, "error", err)
			return nil
		}

		if is_match {
			slog.Info("Match", "path", path)
		}

		return nil
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, uris...)

	if err != nil {
		log.Fatalf("Failed to iterate URIs, %v", err)
	}

}
