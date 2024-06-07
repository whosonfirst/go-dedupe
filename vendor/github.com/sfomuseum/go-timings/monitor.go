package timings

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"io"
	"net/url"
	"sort"
	"strings"
)

// type Monitor provides a common interface for timings-based monitors
type Monitor interface {
	// The Start method starts the monitor dispatching notifications to an io.Writer instance
	Start(context.Context, io.Writer) error
	// The Stop method will stop monitoring
	Stop(context.Context) error
	// The Signal method will dispatch messages to the monitoring process
	Signal(context.Context, ...interface{}) error
}

// monitors is a `aaronland/go-roster.Roster` instance used to maintain a list of registered `Monitor` initialization functions.
var monitors roster.Roster

// MonitorInitializationFunc is a function defined by individual monitor package and used to create
// an instance of that monitor
type MonitorInitializationFunc func(ctx context.Context, uri string) (Monitor, error)

// RegisterMonitor registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Monitor` instances by the `NewMonitor` method.
func RegisterMonitor(ctx context.Context, scheme string, init_func MonitorInitializationFunc) error {

	err := ensureMonitorRoster()

	if err != nil {
		return err
	}

	return monitors.Register(ctx, scheme, init_func)
}

func ensureMonitorRoster() error {

	if monitors == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return fmt.Errorf("Failed to create roster, %w", err)
		}

		monitors = r
	}

	return nil
}

// NewMonitor returns a new `Monitor` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `MonitorInitializationFunc`
// function used to instantiate the new `Monitor`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterMonitor` method.
func NewMonitor(ctx context.Context, uri string) (Monitor, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	scheme := u.Scheme

	i, err := monitors.Driver(ctx, scheme)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve constructor for %s, %w", scheme, err)
	}

	init_func := i.(MonitorInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureMonitorRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range monitors.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
