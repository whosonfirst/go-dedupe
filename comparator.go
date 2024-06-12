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

type Comparator struct {
	database   database.Database
	parser     parser.Parser
	writer     io.Writer
	csv_writer *csvdict.Writer
}

func NewComparator(ctx context.Context, db database.Database, prsr parser.Parser, wr io.Writer) (*Comparator, error) {

	c := &Comparator{
		database: db,
		parser:   prsr,
		writer:   wr,
	}

	return c, nil
}

func (c *Comparator) Compare(ctx context.Context, body []byte, threshold float64) error {

	loc, err := c.parser.Parse(ctx, body)

	if err != nil {
		return fmt.Errorf("Failed to parse feature, %w", err)
	}

	results, err := c.database.Query(ctx, loc.Content(), loc.Metadata())

	if err != nil {
		return fmt.Errorf("Failed to query feature, %w", err)
	}

	for _, qr := range results {

		// slog.Info("Match", "id", "similarity", qr.Similarity, "wof", loc.Content(), "ov", qr.Content)

		if float64(qr.Similarity) >= threshold {
			slog.Info("Match", "similarity", qr.Similarity, "atp", loc.Content(), "ov", qr.Content)

			row := map[string]string{
				"location":   qr.ID,
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
	}

	return nil
}
