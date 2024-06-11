// Packge filters defines interfaces for filtering documents which should be processed during an iteration.
package filters

import (
	"context"
	"io"
)

// type Filters defines an interface for filtering documents which should be processed during an iteration.
type Filters interface {
	// Apply() performs any filtering operations defined by the interface implementation to an `io.ReadSeekCloser` instance and returns a boolean value indicating whether the record should be considered for further processing.
	Apply(context.Context, io.ReadSeeker) (bool, error)
}
