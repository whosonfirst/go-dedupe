package migrate

import (
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
)

var source_repo string
var target_repo string

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("migrate")

	fs.StringVar(&source_repo, "source-repo", "", "The path to the Who's On First repository that deprecated records will be removed from.")
	fs.StringVar(&target_repo, "target-repo", "", "The path to the Who's On First repository that deprecated records will be added from.")

	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Migrate deprecated records from one Who's On First repository to another.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options]", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
