package dedupe

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-dedupe/parser"
	"github.com/whosonfirst/go-dedupe/vector"
)

// Compatator compares arbirtrary locations against a database of existing records.
type Comparator struct {
	location_database     location.Database
	location_parser       parser.Parser
	vector_database_uri   string
	vector_database_cache *ristretto.Cache
	writer                io.Writer
	csv_writer            *csvdict.Writer
	mu                    *sync.RWMutex
}

// ComparatorOptions is a struct containing configuration options used to create a new `Comparator` instance.
type ComparatorOptions struct {
	// LocationDatabaseURI is the URI used to create a (location) `Database` instance of `Location` instances to compare against.
	LocationDatabaseURI string
	// LocationParseURI is the URI used to create a (location) `Parser` instance to derive a `Location` instance from a byte string.
	LocationParserURI string
	// VectorDatabaseURI is the URI used to create `vector.Database` instance used to compare `Location` instances.
	VectorDatabaseURI string
	// Writer is the `io.Writer` instance where CSV rows will be written to.
	Writer io.Writer
}

// NewComparator returns a new `Comparator` instance which wraps all the logic of comparing the embeddings
// for a given `Location` instance against a database of `Location` instances and emit matches as CSV rows.
func NewComparator(ctx context.Context, opts *ComparatorOptions) (*Comparator, error) {

	location_db, err := location.NewDatabase(ctx, opts.LocationDatabaseURI)

	if err != nil {
		return nil, fmt.Errorf("Failed to create location database, %w", err)
	}

	location_parser, err := parser.NewParser(ctx, opts.LocationParserURI)

	if err != nil {
		return nil, fmt.Errorf("Failed to create location parserm %w", err)
	}

	mu := new(sync.RWMutex)

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to create vector database cache, %w", err)
	}

	c := &Comparator{
		location_database:     location_db,
		location_parser:       location_parser,
		vector_database_cache: cache,
		vector_database_uri:   opts.VectorDatabaseURI,
		writer:                opts.Writer,
		mu:                    mu,
	}

	return c, nil
}

// Compare compares 'body' against the database of existing records (contained by 'c'). Matches are written as CSV rows with the
// following keys: location (the location being compared), source (the matching source data that a location is compared against),
// similarity.
func (c *Comparator) Compare(ctx context.Context, body []byte, threshold float64) (bool, error) {

	is_match := false

	loc, err := c.location_parser.Parse(ctx, body)

	if err != nil {
		return is_match, fmt.Errorf("Failed to parse feature, %w", err)
	}

	// Create an in-memory database. This is predicated on the assumption
	// of a limited and manageable number of matches for any given geohash

	geohash := loc.Geohash()

	var vector_db vector.Database

	v, exists := c.vector_database_cache.Get(geohash)

	if !exists {

		db_uri, _ := url.QueryUnescape(c.vector_database_uri)
		db_uri = strings.Replace(db_uri, "{geohash}", geohash, 1)

		new_db, err := vector.NewDatabase(ctx, db_uri)

		if err != nil {
			return false, fmt.Errorf("Failed to create new database, %w", err)
		}

		// c.vector_database_cache.Set(geohash, new_db, 1)
		// c.vector_database_cache.Wait()

		vector_db = new_db
	} else {
		vector_db = v.(vector.Database)
	}

	defer vector_db.Close(ctx)

	count := int32(0)
	t1 := time.Now()

	geohash_cb := func(ctx context.Context, loc *location.Location) error {

		slog.Debug("Add", "geohash", geohash, "loc", loc)
		err := vector_db.Add(ctx, loc)

		if err != nil {
			return fmt.Errorf("Failed to add record, %w", err)
		}

		atomic.AddInt32(&count, 1)
		return nil
	}

	err = c.location_database.GetWithGeohash(ctx, geohash, geohash_cb)

	if err != nil {
		return false, fmt.Errorf("Failed to retrieve records for geohash, %w", err)
	}

	if atomic.LoadInt32(&count) == 0 {
		return is_match, nil
	}

	slog.Debug("Candidates", "geohash", geohash, "count", atomic.LoadInt32(&count), "time", time.Since(t1))

	results, err := vector_db.Query(ctx, loc)

	if err != nil {
		slog.Error("Failed to query", "geohash", geohash, "count", atomic.LoadInt32(&count), "error", err)
		return is_match, fmt.Errorf("Failed to query feature, %w", err)
	}

	// slog.Info("Possible", "geohash", geohash, "count", len(results))

	// END OF new new ....

	for _, qr := range results {

		slog.Info("Possible", "geohash", geohash, "similarity", qr.Similarity, "wof", loc.String(), "ov", qr.Content)

		// Make this a toggle...
		if float64(qr.Similarity) == 0 || float64(qr.Similarity) <= threshold {

			slog.Info("Match", "geohash", geohash, "threshold", threshold, "similarity", qr.Similarity, "query", loc.String(), "candidate", qr.Content)
			is_match = true

			row := map[string]string{
				// The location being compared
				"location": qr.ID,
				// The matching source data that a location is compared against
				"source":     loc.ID,
				"similarity": fmt.Sprintf("%02f", qr.Similarity),
			}

			c.mu.Lock()
			defer c.mu.Unlock()

			if c.csv_writer == nil {

				fieldnames := make([]string, 0)

				for k, _ := range row {
					fieldnames = append(fieldnames, k)
				}

				wr, err := csvdict.NewWriter(c.writer, fieldnames)

				if err != nil {
					return is_match, fmt.Errorf("Failed to create CSV writer, %w", err)
				}

				err = wr.WriteHeader()

				if err != nil {
					return is_match, fmt.Errorf("Failed to write header for CSV writer, %w", err)
				}

				c.csv_writer = wr
			}

			err = c.csv_writer.WriteRow(row)

			if err != nil {
				return is_match, fmt.Errorf("Failed to write header for CSV writer, %w", err)
			}

			break
		}
	}

	c.Flush()

	return is_match, nil
}

func (c *Comparator) Flush() {

	if c.csv_writer != nil {
		c.csv_writer.Flush()
	}
}

func (c *Comparator) Close() {

	// vector_database_cache
}
