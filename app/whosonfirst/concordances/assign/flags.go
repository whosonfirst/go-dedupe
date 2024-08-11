package assign

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("assign")

	fs.BoolVar(&verbose, "verbose", false, "...")
	return fs
}
