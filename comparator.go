package dedupe

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
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
	location_database         location.Database
	location_parser           parser.Parser
	embeddings_database_uri   string
	embeddings_database_cache *ristretto.Cache
	writer                    io.Writer
	csv_writer                *csvdict.Writer
	mu                        *sync.RWMutex
}

type ComparatorOptions struct {
	LocationDatabaseURI   string
	LocationParserURI     string
	EmbeddingsDatabaseURI string
	Writer                io.Writer
}

// NewComparator returns a new `Comparator` instance. 'db' is the `database.Database` instance of existing records to compare
// locations against, 'prsr' is the `parser.Parser` instance to convert a location in to a `parser.Location` instance and `wr'
// is a `io.Writer` instance where match results will be written.
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
		return nil, err
	}

	c := &Comparator{
		location_database:         location_db,
		location_parser:           location_parser,
		embeddings_database_cache: cache,
		embeddings_database_uri:   opts.EmbeddingsDatabaseURI,
		writer:                    opts.Writer,
		mu:                        mu,
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

	// START OF new new ....

	// Create an in-memory database. This is predicated on the assumption
	// of a limited and manageable number of matches for any given geohash

	geohash := loc.Geohash()

	var embeddings_db database.Database

	v, exists := c.embeddings_database_cache.Get(geohash)

	if !exists {

		db_uri := c.embeddings_database_uri
		db_uri = strings.Replace(db_uri, "{geohash}", geohash, 1)

		new_db, err := database.NewDatabase(ctx, db_uri)

		if err != nil {
			return false, fmt.Errorf("Failed to create new database, %w", err)
		}

		c.embeddings_database_cache.Set(geohash, new_db, 1)
		c.embeddings_database_cache.Wait()

		embeddings_db = new_db
	} else {
		embeddings_db = v.(database.Database)
	}

	count := int32(0)
	// t1 := time.Now()

	geohash_cb := func(ctx context.Context, loc *location.Location) error {

		// slog.Info("Add", "geohash", geohash, "loc", loc)
		err := embeddings_db.Add(ctx, loc)

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

	// slog.Info("Candidates", "geohash", geohash, "count", atomic.LoadInt32(&count), "time", time.Since(t1))

	/*

			Do not understand these errors:

		2024/07/05 13:21:29 ERROR Failed to query geohash=u0myt count=3 error="Failed to query, nResults must be <= the number of documents in the collection"
		2024/07/05 13:21:29 WARN Failed to compare feature path=/usr/local/data/alltheplaces/aldi_sud_de.geojson error="Failed to query feature, Failed to query, nResults must be <= the number of documents in the collection"
		2024/07/05 13:22:11 INFO Matches path=/usr/local/data/alltheplaces/aldi_sud_de.geojson features=2019 matches=430 "total features"=48678 "total matches"=1568

	*/

	results, err := embeddings_db.Query(ctx, loc)

	if err != nil {
		slog.Error("Failed to query", "geohash", geohash, "count", atomic.LoadInt32(&count), "error", err)
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
