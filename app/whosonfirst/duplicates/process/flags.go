package process

import (
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
)

var reader_uri string
var writer_uri string

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("process")

	fs.StringVar(&reader_uri, "reader-uri", "", "A valid whosonfirst/go-reader.Reader URI that records to be processed will be read from.")
	fs.StringVar(&writer_uri, "writer-uri", "stdout://", "A valid whosonfirst/go-writer.Writer URI where updated records will be written to.")

	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Process duplicate records in a Who's On First repository (which means deprecate and mark as superseding or superseded by where necessary).\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] uri(N) uri(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
