package assign

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var reader_uri string
var writer_uri string

var concordance_namespace string
var concordance_predicate string
var concordance_as_int bool
var wof_label string

var mark_is_current bool

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("assign")

	fs.StringVar(&reader_uri, "reader-uri", "", "A valid whosonfirst/go-reader URI for reading WOF records from.")
	fs.StringVar(&writer_uri, "writer-uri", "", "A valid whosonfirst/go-reader URI for writing WOF records from.")

	fs.StringVar(&wof_label, "whosonfirst-label", "target", "The \"label\" used to identify WOF records. Valid options are: source, target.")

	fs.StringVar(&concordance_namespace, "concordance-namespace", "", "The namespace of the concordance being applied.")
	fs.StringVar(&concordance_predicate, "concordance-predicate", "id", "The predicate of the concordance being applies.")
	fs.BoolVar(&concordance_as_int, "concordance-as-int", false, "If true cast the concordance ID as an int64")

	fs.BoolVar(&mark_is_current, "mark-is-current", false, "If true the addition of a cocordance will mark this record as mz:is_current=1")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")
	return fs
}
