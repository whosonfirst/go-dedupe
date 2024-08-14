package compare

import (
	"flag"
	"fmt"
	"os"

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

	fs.StringVar(&vector_database_uri, "vector-database-uri", "sqlite://?model=mxbai-embed-large&dsn=%7Btmp%7D%7Bgeohash%7D.db%3Fcache%3Dshared%26mode%3Dmemory&embedder-uri=ollama%3A%2F%2F%3Fmodel%3Dmxbai-embed-large&max-distance=4&max-results=10&dimensions=1024&compression=none", "A valid whosonfirst/go-dedupe/vector.Database URI.")

	fs.StringVar(&source_location_database_uri, "source-location-database-uri", "", "A valid whosonfirst/go-dedupe/location.Database URI.")
	fs.StringVar(&target_location_database_uri, "target-location-database-uri", "", "A valid whosonfirst/go-dedupe/location.Database URI.")

	fs.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "A valid sfomuseum/go-timings.Monitor URI.")

	fs.Float64Var(&threshold, "threshold", 4.0, "The threshold value for matching records. Whether this value is greater than or lesser than a matching value will be dependent on the vector database in use.")

	fs.IntVar(&workers, "workers", 10, "The number of simultaneous worker processes to use.")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Compare two location databases and emit matching records as CSV-encoded rows.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
