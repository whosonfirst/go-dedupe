package assign

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var reader_uri string
var writer_uri string

var concordance_namespace string
var concordance_predicate string
var wof_label string

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("assign")

	fs.StringVar(&reader_uri, "reader-uri", "", "...")
	fs.StringVar(&writer_uri, "writer-uri", "", "...")

	fs.StringVar(&wof_label, "whosonfirst-label", "target", "...")

	fs.StringVar(&concordance_namespace, "concordance-namespace", "", "...")
	fs.StringVar(&concordance_predicate, "concordance-predicate", "", "...")

	fs.BoolVar(&verbose, "verbose", false, "...")
	return fs
}
