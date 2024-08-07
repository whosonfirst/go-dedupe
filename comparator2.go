package dedupe

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	_ "sync/atomic"
	"time"

	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/parser"
	"github.com/whosonfirst/go-dedupe/vector"
)

// Compatator compares arbirtrary locations against a database of existing records.
type Comparator2 struct {
	source_database     location.Database
	target_database     location.Database
	vector_database_uri string
	writer              io.Writer
	csv_writer          *csvdict.Writer
	mu                  *sync.RWMutex
}

// ComparatorOptions is a struct containing configuration options used to create a new `Comparator` instance.
type Comparator2Options struct {
	// LocationDatabaseURI is the URI used to create a (location) `Database` instance of `Location` instances to compare against.
	SourceLocationDatabaseURI string
	TargetLocationDatabaseURI string
	// VectorDatabaseURI is the URI used to create `vector.Database` instance used to compare `Location` instances.
	VectorDatabaseURI string
	// Writer is the `io.Writer` instance where CSV rows will be written to.
	Writer io.Writer
}

// NewComparator returns a new `Comparator` instance which wraps all the logic of comparing the embeddings
// for a given `Location` instance against a database of `Location` instances and emit matches as CSV rows.
func NewComparator2(ctx context.Context, opts *Comparator2Options) (*Comparator2, error) {

	source_db, err := location.NewDatabase(ctx, opts.SourceLocationDatabaseURI)

	if err != nil {
		return nil, fmt.Errorf("Failed to create location database, %w", err)
	}

	target_db, err := location.NewDatabase(ctx, opts.TargetLocationDatabaseURI)

	if err != nil {
		return nil, fmt.Errorf("Failed to create location database, %w", err)
	}

	mu := new(sync.RWMutex)

	c := &Comparator2{
		source_database:     source_db,
		target_database:     target_db,
		vector_database_uri: opts.VectorDatabaseURI,
		writer:              opts.Writer,
		mu:                  mu,
	}

	return c, nil
}

// Compare compares 'body' against the database of existing records (contained by 'c'). Matches are written as CSV rows with the
// following keys: location (the location being compared), source (the matching source data that a location is compared against),
// similarity.
func (c *Comparator2) Compare(ctx context.Context, threshold float64) error {

	// For each geohash in the target database

	geohashes_cb := func(ctx context.Context, geohash string) error {

		slog.Debug("Process geohash", "geohash", geohash)

		// Create vector database for geohash

		db_uri, _ := url.QueryUnescape(c.vector_database_uri)
		db_uri = strings.Replace(db_uri, "{geohash}", geohash, 1)

		vector_db, err := vector.NewDatabase(ctx, db_uri)

		if err != nil {
			return fmt.Errorf("Failed to create new database, %w", err)
		}

		defer vector_db.Close(ctx)

		// Index vector database with records matching geohash in source database

		add_cb := func(ctx context.Context, loc *location.Location) error {

			slog.Debug("Add", "geohash", geohash, "loc", loc)
			err := vector_db.Add(ctx, loc)

			if err != nil {
				return fmt.Errorf("Failed to add record, %w", err)
			}

			// atomic.AddInt32(&count, 1)
			return nil
		}

		t1 := time.Now()

		slog.Debug("Get locations with geohash from source database", "geohash", geohash)
		err = c.source_database.GetWithGeohash(ctx, geohash, add_cb)

		if err != nil {
			return fmt.Errorf("Failed to add source locations to vector database, %w", err)
		}

		slog.Debug("Time to add locations with geohash from source database", "geohash", geohash, "time", time.Since(t1))

		// Get all the records matching geohash in target database and compare each against vector database

		compare_cb := func(ctx context.Context, loc *location.Location) error {

			slog.Info("Compare location from target database", "geohash", geohash, "location", loc.String())

			results, err := vector_db.Query(ctx, loc)

			if err != nil {
				slog.Error("Failed to query", "geohash", geohash, "location", loc.String(), "error", err)
				return fmt.Errorf("Failed to query feature, %w", err)
			}

			for _, qr := range results {

				slog.Info("Possible", "geohash", geohash, "similarity", qr.Similarity, "wof", loc.String(), "ov", qr.Content)

				// Make this a toggle...
				if float64(qr.Similarity) > threshold {
					continue
				}

				slog.Info("Match", "geohash", geohash, "threshold", threshold, "similarity", qr.Similarity, "query", loc.String(), "candidate", qr.Content)

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
						return fmt.Errorf("Failed to create CSV writer, %w", err)
					}

					err = wr.WriteHeader()

					if err != nil {
						return fmt.Errorf("Failed to write header for CSV writer, %w", err)
					}

					c.csv_writer = wr
				}

				err = c.csv_writer.WriteRow(row)

				if err != nil {
					return fmt.Errorf("Failed to write header for CSV writer, %w", err)
				}

				break
			}

			return nil
		}

		slog.Debug("Get locations with geohash from target database", "geohash", geohash)
		err = c.target_database.GetWithGeohash(ctx, geohash, compare_cb)

		if err != nil {
			return err
		}

		return nil
	}

	slog.Debug("Get geohashes from target database")
	err := c.target_database.GetGeohashes(ctx, geohashes_cb)

	if err != nil {
		return err
	}

	c.Flush()

	return nil
}

func (c *Comparator2) Flush() {

	if c.csv_writer != nil {
		c.csv_writer.Flush()
	}
}

func (c *Comparator2) Close() {

}
