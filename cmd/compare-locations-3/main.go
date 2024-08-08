package main

/*

[asc][asc@SD-931-4][11:05:22] /usr/local/whosonfirst/go-whosonfirst-dedupe                                                                                                                     > go run cmd/compare-alltheplaces/main.go /usr/local/data/alltheplaces/dunkin_us.geojson

 51422 |2024/08/06 19:30:04 INFO Unmarshal path=/usr/local/data/alltheplaces/amcal_au.geojson
 51423 |2024/08/06 19:30:08 INFO Possible geohash=qd4qe similarity=3.790432929992676 wof="Amcal+ Pharmacy Ravenswood, 3/60 Lloyd Avenue Ravenswood WA AU" ov="Ravenswood Amcal Pharmacy, 60 Lloyd Av\
       |e Ravenswood WA AU"
 51424 |2024/08/06 19:30:08 INFO Match geohash=qd4qe similarity=3.790432929992676 atp="Amcal+ Pharmacy Ravenswood, 3/60 Lloyd Avenue Ravenswood WA AU" ov="Ravenswood Amcal Pharmacy, 60 Lloyd Ave R\
       |avenswood WA AU"
 51425 |2024/08/06 19:30:11 INFO Matches path=/usr/local/data/alltheplaces/amcal_au.geojson features=210 matches=1 "total features"=65582 "total matches"=94

*/

import (
	"context"
	"flag"
	"log"
	"log/slog"

	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	"github.com/whosonfirst/go-dedupe/compare"
	_ "github.com/whosonfirst/go-dedupe/overture"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	var vector_database_uri string

	var source_location_database_uri string
	var target_location_database_uri string

	var monitor_uri string
	var workers int

	var threshold float64
	var verbose bool

	flag.StringVar(&vector_database_uri, "vector-database-uri", "sqlite://?model=mxbai-embed-large&dsn=%7Btmp%7D%7Bgeohash%7D.db%3Fcache%3Dshared%26mode%3Dmemory&embedder-uri=ollama%3A%2F%2F%3Fmodel%3Dmxbai-embed-large&max-distance=4&max-results=10&dimensions=1024&compression=none", "...")

	flag.StringVar(&source_location_database_uri, "source-location-database-uri", "sql://sqlite3?dsn=/usr/local/data/overture/overture-locations.db", "...")
	flag.StringVar(&target_location_database_uri, "target-location-database-uri", "sql://sqlite3?dsn=/usr/local/data/overture/whosonfirst-locations.db", "...")

	flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")

	flag.Float64Var(&threshold, "threshold", 4.0, "...")

	flag.IntVar(&workers, "workers", 10, "...")
	flag.BoolVar(&verbose, "verbose", false, "...")
	flag.Parse()

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()

	cmp_opts := &compare.CompareLocationDatabasesOptions{
		SourceLocationDatabaseURI: source_location_database_uri,
		TargetLocationDatabaseURI: target_location_database_uri,
		VectorDatabaseURI:         vector_database_uri,
		Threshold:                 threshold,
	}

	err := compare.CompareLocationDatabases(ctx, cmp_opts)

	if err != nil {
		log.Fatalf("Failed to create new comparator, %v", err)
	}
}
