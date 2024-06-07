package timings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"sync"
	"time"
)

type SinceEvent uint8

const (
	SinceStart SinceEvent = iota
	SinceStop
)

// type SinceMonitor implements the `Monitor` interface providing a background timings mechanism that tracks the duration of time
// between events.
type SinceMonitor struct {
	Monitor
	done_ch  chan bool
	since_ch chan *SinceResponse
	start    time.Time
	events   *sync.Map
	mu       *sync.RWMutex
}

// SinceResponse is a struct containing information related to a "since" timing event.
type SinceResponse struct {
	// Label is a string that was included with a `Signal` event
	Label string `json:"message"`
	// Duration is the string representation of a `time.Duuration` which is the amount of time that elapsed between `Signal` events
	Duration string `json:"duration"`
	// Timestamp is the Unix timestamp when the `SinceResponse` was created
	Timestamp int64 `json:"timestamp"`
}

func init() {
	ctx := context.Background()
	RegisterMonitor(ctx, "since", NewSinceMonitor)
}

// String() returns a string representation of the response.
func (s SinceResponse) String() string {
	return fmt.Sprintf("%d %s %s", s.Timestamp, s.Label, s.Duration)
}

// NewSinceMonitor creates a new `SinceMonitor` instance that will dispatch notifications using a time.Ticker configured
// by 'uri' which is expected to take the form of:
//
//	since://
func NewSinceMonitor(ctx context.Context, uri string) (Monitor, error) {

	_, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	done_ch := make(chan bool)
	since_ch := make(chan *SinceResponse)

	events := new(sync.Map)
	mu := new(sync.RWMutex)

	t := &SinceMonitor{
		done_ch:  done_ch,
		since_ch: since_ch,
		mu:       mu,
		events:   events,
	}

	return t, nil
}

// Start() will cause background monitoring to begin, dispatching notifications to wr in
// the form of JSON-encoded `SinceResponse` values.
func (t *SinceMonitor) Start(ctx context.Context, wr io.Writer) error {

	if !t.start.IsZero() {
		return fmt.Errorf("Monitor has already been started")
	}

	now := time.Now()
	t.start = now

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.done_ch:
				return
			case rsp := <-t.since_ch:

				enc := json.NewEncoder(wr)
				err := enc.Encode(rsp)

				if err != nil {
					log.Printf("Failed to encode response, %v", err)
				}
			}
		}
	}()

	return nil
}

// Stop() will cause background monitoring to be halted.
func (t *SinceMonitor) Stop(ctx context.Context) error {

	t.events.Range(func(k interface{}, v interface{}) bool {

		label := k.(string)
		last_event := v.(time.Time)

		duration := time.Since(last_event)
		now := time.Now()

		rsp := &SinceResponse{
			Label:     label,
			Timestamp: now.Unix(),
			Duration:  duration.String(),
		}

		t.since_ch <- rsp

		t.events.Delete(k)
		return true
	})

	t.done_ch <- true
	return nil
}

// Signal will cause the background monitors since to be incremented by one.
func (t *SinceMonitor) Signal(ctx context.Context, args ...interface{}) error {

	if len(args) < 2 {
		return fmt.Errorf("Signal requires valid status event and label")
	}

	switch args[0].(type) {
	case SinceEvent:
		// pass
	default:
		return fmt.Errorf("Signal requires first argument to be a SinceEvent")
	}

	switch args[1].(type) {
	case string:
		// pass
	default:
		return fmt.Errorf("Signal requires second argument to a string")
	}

	ev := args[0].(SinceEvent)
	label := args[1].(string)

	var last_event time.Time

	switch ev {
	case SinceStart:

		now := time.Now()

		_, ok := t.events.LoadOrStore(label, now)

		if ok {
			return fmt.Errorf("Label %s has already been stored", label)
		}

		return nil

	case SinceStop:

		v, ok := t.events.LoadAndDelete(label)

		if !ok {
			return fmt.Errorf("Failed to find any record with label %s", label)
		}

		last_event = v.(time.Time)

	default:
		return fmt.Errorf("Unsupported event")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	duration := time.Since(last_event)

	rsp := &SinceResponse{
		Label:     label,
		Timestamp: now.Unix(),
		Duration:  duration.String(),
	}

	t.since_ch <- rsp
	return nil
}
