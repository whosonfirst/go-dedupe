package timings

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/sfomuseum/iso8601duration"
)

// type CounterMonitor implements the `Monitor` interface providing a background timings mechanism that tracks incrementing events.
type CounterMonitor struct {
	Monitor
	done_ch chan bool
	start   time.Time
	counter int64
	ticker  *time.Ticker
	total   int64
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

	q := u.Query()

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

	if q.Has("total") {

		v, err := strconv.ParseInt(q.Get("total"), 10, 64)

		if err != nil {
			return nil, fmt.Errorf("Invalid ?total= query parameter, %w", err)
		}

		t.total = v
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

	logger := slog.New(slog.NewTextHandler(wr, nil))
	
	go func() {

		for {
			select {
			case <-t.done_ch:
				return
			case <-ctx.Done():
				return
			case <-t.ticker.C:

				var msg string

				if t.total > 0 {
					msg = fmt.Sprintf("Processed %d/%d records in %v (started %v)", atomic.LoadInt64(&t.counter), t.total, time.Since(t.start), t.start)
				} else {
					msg = fmt.Sprintf("Processed %d records in %v (started %v)", atomic.LoadInt64(&t.counter), time.Since(t.start), t.start)
				}

				logger.Info(msg)
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
