package migrate

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var source_repo string
var target_repo string

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("migrate")

	fs.StringVar(&source_repo, "source-repo", "", "...")
	fs.StringVar(&target_repo, "target-repo", "", "...")

	fs.BoolVar(&verbose, "verbose", false, "...")
	return fs
}
