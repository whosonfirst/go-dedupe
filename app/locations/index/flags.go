package index

import (
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
)

var location_database_uri string
var location_parser_uri string
var iterator_uri string

var monitor_uri string

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("index")

	fs.StringVar(&location_database_uri, "location-database-uri", "", "A valid whosonfirst/go-dedupe/location.Database URI.")
	fs.StringVar(&location_parser_uri, "location-parser-uri", "", "A valid whosonfirst/go-dedupe/location.Parser URI.")
	fs.StringVar(&iterator_uri, "iterator-uri", "", "A valid whosonfirst/go-dedupe/iterator.Iterator URI.")

	fs.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "A valid sfomuseum/go-timings.Monitor URI.")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Populate (index) a location database from data/provider source..\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] uri(N) uri(N)", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
