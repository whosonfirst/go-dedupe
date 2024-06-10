package writer

import (
	"io"
	"log"
)

// DefaultLogger() returns a `log.Logger` instance that writes to `io.Discard`.
func DefaultLogger() *log.Logger {
	return log.New(io.Discard, "", log.Lshortfile)
}
