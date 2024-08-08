package compare

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var vector_database_uri string

var source_location_database_uri string
var target_location_database_uri string

var monitor_uri string
var workers int

var threshold float64
var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("compare")

	fs.StringVar(&vector_database_uri, "vector-database-uri", "sqlite://?model=mxbai-embed-large&dsn=%7Btmp%7D%7Bgeohash%7D.db%3Fcache%3Dshared%26mode%3Dmemory&embedder-uri=ollama%3A%2F%2F%3Fmodel%3Dmxbai-embed-large&max-distance=4&max-results=10&dimensions=1024&compression=none", "...")

	fs.StringVar(&source_location_database_uri, "source-location-database-uri", "sql://sqlite3?dsn=/usr/local/data/overture/alltheplaces-locations.db", "...")
	fs.StringVar(&target_location_database_uri, "target-location-database-uri", "sql://sqlite3?dsn=/usr/local/data/overture/whosonfirst-locations.db", "...")

	fs.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")

	fs.Float64Var(&threshold, "threshold", 4.0, "...")

	fs.IntVar(&workers, "workers", 10, "...")
	fs.BoolVar(&verbose, "verbose", false, "...")

	return fs
}
