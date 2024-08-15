package compare

import (
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
)

var vector_database_uri string

var vector_database_dsn string
var vector_database_embedder_uri string
var vector_database_model string

var source_location_database_uri string
var target_location_database_uri string

var monitor_uri string
var workers int

var threshold float64
var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("compare")

	fs.StringVar(&vector_database_uri, "vector-database-uri", "sqlite://?model={vector-database-model}&dsn={vector-database-dsn}&embedder-uri={vector-database-embedder-uri}&max-distance=4&max-results=10&dimensions=1024&compression=none", "A valid whosonfirst/go-dedupe/vector.Database URI.")

	fs.StringVar(&vector_database_dsn, "vector-database-dsn", "{tmp}{geohash}.db?cache=shared&mode=memory", "A valid whosonfirst/go-dedupe/vector.Database DSN string. If the parameter contains the string \"{geohash}\" then that string will be replaced, at runtime, with the value of the geohash being compared. This will have the effect of creating a vector database per geohash. This value will be used to replace any \"{vector-database-dsn}\" strings in the -vector-database-uri flag.")

	fs.StringVar(&vector_database_embedder_uri, "vector-database-embedder-uri", "ollama://?model={vector-database-model}", "A valid whosonfirst/go-dedupe/embeddings.Embedder URI. This value will be used to replace any \"{vector-database-embedder-uri}\" strings in the -vector-database-uri flag.")

	fs.StringVar(&vector_database_model, "vector-database-model", "mxbai-embed-large", "The name of the model to use comparing records in the location database against records in the vector database. This value will be used to replace any \"{vector-database-model}\" strings in the -vector-database-uri and -vector-database-embedder-uri flags.")

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
