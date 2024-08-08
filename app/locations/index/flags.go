package index

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var location_database_uri string
var location_parser_uri string
var iterator_uri string

var monitor_uri string

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("index")

	fs.StringVar(&location_database_uri, "location-database-uri", "", "...")
	fs.StringVar(&location_parser_uri, "location-parser-uri", "", "...")
	fs.StringVar(&iterator_uri, "iterator-uri", "", "...")

	fs.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")
	fs.BoolVar(&verbose, "verbose", false, "...")

	return fs
}
