package timings

import (
	"context"
	"fmt"
	"github.com/sfomuseum/iso8601duration"
	"io"
	_ "log"
	"net/url"
	"sync/atomic"
	"time"
)

// type CounterMonitor implements the `Monitor` interface providing a background timings mechanism that tracks incrementing events.
type CounterMonitor struct {
	Monitor
	done_ch chan bool
	start   time.Time
	counter int64
	ticker  *time.Ticker
}

func init() {
	ctx := context.Background()
	RegisterMonitor(ctx, "counter", NewCounterMonitor)
}

// NewCounterMonitor creates a new `CounterMonitor` instance that will dispatch notifications using a time.Ticker configured
// by 'uri' which is expected to take the form of:
//
//	counter://?duration={ISO8601_DURATION}
//
// Where {ISO8601_DURATION} is a valid ISO8601 duration string.
func NewCounterMonitor(ctx context.Context, uri string) (Monitor, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	str_duration := u.Host

	if str_duration == "" {
		return nil, fmt.Errorf("Missing duration parameter, %w", err)
	}

	d, err := duration.FromString(str_duration)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse '%s', %w", str_duration, err)
	}

	done_ch := make(chan bool)
	count := int64(0)

	ticker := time.NewTicker(d.ToDuration())

	t := &CounterMonitor{
		done_ch: done_ch,
		ticker:  ticker,
		counter: count,
	}

	return t, nil
}

// Start() will cause background monitoring to begin, dispatching notifications to wr.
func (t *CounterMonitor) Start(ctx context.Context, wr io.Writer) error {

	if !t.start.IsZero() {
		return fmt.Errorf("Monitor has already been started")
	}

	now := time.Now()
	t.start = now

	go func() {

		for {
			select {
			case <-t.done_ch:
				return
			case <-ctx.Done():
				return
			case <-t.ticker.C:
				msg := fmt.Sprintf("processed %d records in %v (started %v)\n", atomic.LoadInt64(&t.counter), time.Since(t.start), t.start)
				wr.Write([]byte(msg))
			}
		}
	}()

	return nil
}

// Stop() will cause background monitoring to be halted.
func (t *CounterMonitor) Stop(ctx context.Context) error {
	t.done_ch <- true
	return nil
}

// Signal will cause the background monitors counter to be incremented by one.
func (t *CounterMonitor) Signal(ctx context.Context, args ...interface{}) error {
	return t.increment(ctx)
}

func (t *CounterMonitor) increment(ctx context.Context) error {
	atomic.AddInt64(&t.counter, 1)
	return nil
}
