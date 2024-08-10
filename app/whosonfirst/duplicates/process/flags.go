package process

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var reader_uri string
var writer_uri string

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("process")

	fs.StringVar(&reader_uri, "reader-uri", "", "...")
	fs.StringVar(&writer_uri, "writer-uri", "stdout://", "...")

	fs.BoolVar(&verbose, "verbose", false, "...")
	return fs
}
