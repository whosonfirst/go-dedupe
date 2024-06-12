package dedupe

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-dedupe/database"
	"github.com/whosonfirst/go-dedupe/parser"
)

// Compatator compares arbirtrary locations against a database of existing records.
type Comparator struct {
	database   database.Database
	parser     parser.Parser
	writer     io.Writer
	csv_writer *csvdict.Writer
}

// NewComparator returns a new `Comparator` instance. 'db' is the `database.Database` instance of existing records to compare
// locations against, 'prsr' is the `parser.Parser` instance to convert a location in to a `parser.Location` instance and `wr'
// is a `io.Writer` instance where match results will be written.
func NewComparator(ctx context.Context, db database.Database, prsr parser.Parser, wr io.Writer) (*Comparator, error) {

	c := &Comparator{
		database: db,
		parser:   prsr,
		writer:   wr,
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

	results, err := c.database.Query(ctx, loc)

	if err != nil {
		return is_match, fmt.Errorf("Failed to query feature, %w", err)
	}

	for _, qr := range results {

		// slog.Info("Match", "id", "similarity", qr.Similarity, "wof", loc.Content(), "ov", qr.Content)

		if float64(qr.Similarity) >= threshold {

			slog.Info("Match", "similarity", qr.Similarity, "atp", loc.String(), "ov", qr.Content)
			is_match = true

			row := map[string]string{
				// The location being compared
				"location": qr.ID,
				// The matching source data that a location is compared against
				"source":     loc.ID,
				"similarity": fmt.Sprintf("%02f", qr.Similarity),
			}

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

	return is_match, nil
}

func (c *Comparator) Flush() {

	if c.csv_writer != nil {
		c.csv_writer.Flush()
	}
}
