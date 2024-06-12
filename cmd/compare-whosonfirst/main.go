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
	"github.com/whosonfirst/go-dedupe/database"
	_ "github.com/whosonfirst/go-dedupe/overture"
	"github.com/whosonfirst/go-dedupe/parser"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

func main() {

	var database_uri string
	var parser_uri string
	var iterator_uri string
	var threshold float64

	flag.StringVar(&database_uri, "database-uri", "opensearch://?dsn=https%3A%2F%2Flocalhost%3A9200%2Fdedupe%3Fusername%3Dadmin%26password%3DKJHFGDFJGSJfsdkjfhsdoifruwo45978h52dcn%26insecure%3Dtrue%26require-tls%3Dtrue&model=a8-aBJABf__qJekL_zJC&bulk-index=false", "...")

	//flag.StringVar(&database_uri, "database-uri", "chromem://venues/usr/local/data/venues.db?model=mxbai-embed-large", "...")
	flag.StringVar(&parser_uri, "parser-uri", "whosonfirstvenues://", "...")
	flag.StringVar(&iterator_uri, "iterator-uri", "repo://", "...")

	flag.Float64Var(&threshold, "threshold", 0.75, "...")
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

	cmp, err := dedupe.NewComparator(ctx, db, prsr, os.Stdout)

	if err != nil {
		log.Fatalf("Failed to create new comparator, %v", err)
	}

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %v", path, err)
		}

		err = cmp.Compare(ctx, body, threshold)

		if err != nil {
			slog.Warn("Failed to compare feature", "path", path, "error", err)
			return nil
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
