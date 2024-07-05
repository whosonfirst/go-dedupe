package dedupe

import (
	"context"
	"fmt"
	"io"
	_ "log/slog"
	"sync"
	"sync/atomic"
	_ "time"

	"github.com/dgraph-io/ristretto"
	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-dedupe/database"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-dedupe/parser"
)

// Compatator compares arbirtrary locations against a database of existing records.
type Comparator struct {
	location_database location.Database
	database          database.Database
	database_cache    *ristretto.Cache
	parser            parser.Parser
	writer            io.Writer
	csv_writer        *csvdict.Writer
	mu                *sync.RWMutex
}

// NewComparator returns a new `Comparator` instance. 'db' is the `database.Database` instance of existing records to compare
// locations against, 'prsr' is the `parser.Parser` instance to convert a location in to a `parser.Location` instance and `wr'
// is a `io.Writer` instance where match results will be written.
func NewComparator(ctx context.Context, db location.Database, prsr parser.Parser, wr io.Writer) (*Comparator, error) {

	mu := new(sync.RWMutex)

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	if err != nil {
		return nil, err
	}

	c := &Comparator{
		location_database: db,
		database_cache:    cache,
		parser:            prsr,
		writer:            wr,
		mu:                mu,
	}

	return c, nil
}

// Compare compares 'body' against the database of existing records (contained by 'c'). Matches are written as CSV rows with the
// following keys: location (the location being compared), source (the matching source data that a location is compared against),
// similarity.
func (c *Comparator) Compare(ctx context.Context, body []byte, threshold float64) (bool, error) {

	is_match := false

	loc, err := c.parser.Parse(ctx, body)

	if err != nil {
		return is_match, fmt.Errorf("Failed to parse feature, %w", err)
	}

	/*
		results, err := c.database.Query(ctx, loc)

		if err != nil {
			return is_match, fmt.Errorf("Failed to query feature, %w", err)
		}
	*/

	// START OF new new ....

	// Create an in-memory database. This is predicated on the assumption
	// of a limited and manageable number of matches for any given geohash

	geohash := loc.Geohash()

	var db database.Database

	v, exists := c.database_cache.Get(geohash)

	if !exists {

		db_uri := fmt.Sprintf("chromem://%s?model=mxbai-embed-large", geohash)
		new_db, err := database.NewDatabase(ctx, db_uri)

		if err != nil {
			return false, fmt.Errorf("Failed to create new database, %w", err)
		}

		c.database_cache.Set(geohash, new_db, 1)
		c.database_cache.Wait()
		db = new_db
	} else {
		db = v.(database.Database)
	}

	count := int32(0)
	// t1 := time.Now()

	geohash_cb := func(ctx context.Context, loc *location.Location) error {

		// slog.Info("Add", "geohash", geohash, "loc", loc)
		err := db.Add(ctx, loc)

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

	// slog.Info("Candidates", "geohash", geohash, "count", atomic.LoadInt32(&count), "time", time.Since(t1))

	results, err := db.Query(ctx, loc)

	if err != nil {
		return is_match, fmt.Errorf("Failed to query feature, %w", err)
	}

	// slog.Info("Possible", "geohash", geohash, "count", len(results))

	// END OF new new ....

	for _, qr := range results {

		// slog.Info("Match", "id", "similarity", qr.Similarity, "wof", loc.Content(), "ov", qr.Content)

		if float64(qr.Similarity) >= threshold {

			// slog.Info("Match", "similarity", qr.Similarity, "atp", loc.String(), "ov", qr.Content)
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
