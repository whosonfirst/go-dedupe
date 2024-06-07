package timings

import (
	"context"
	"io"
)

// type NullMonitor implements the `Monitor` interface but does nothing.
type NullMonitor struct {
	Monitor
}

func init() {
	ctx := context.Background()
	RegisterMonitor(ctx, "null", NewNullMonitor)
}

// NewNullMonitor() creates a new `NullMonitor` which does nothing.
//
//	null://
func NewNullMonitor(ctx context.Context, uri string) (Monitor, error) {
	nm := &NullMonitor{}
	return nm, nil
}

// Start() will cause monitoring to begin, which in this means nothing will happen and
// no events will be written to 'wr'.
func (nm *NullMonitor) Start(ctx context.Context, wr io.Writer) error {
	return nil
}

// Stop() will cause monitoring to be halted.
func (nm *NullMonitor) Stop(ctx context.Context) error {
	return nil
}

// Signal() is a no-op and nothing will happen.
func (nm *NullMonitor) Signal(ctx context.Context, args ...interface{}) error {
	return nil
}
