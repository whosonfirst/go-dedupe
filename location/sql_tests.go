package location

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/paulmach/orb"
)

func testSQLDatabaseEngine(ctx context.Context, engine string) error {

	suffix := fmt.Sprintf("*-%s.db", engine)
	f, err := os.CreateTemp("", suffix)

	if err != nil {
		return fmt.Errorf("Failed to create temp file for %s, %v", suffix, err)
	}

	tmp_path := f.Name()

	// Close and remove the temp file so the database/sql driver will create it
	// from scratch

	err = f.Close()

	if err != nil {
		return fmt.Errorf("Failed to close temp file, %w", err)
	}

	err = os.Remove(tmp_path)

	if err != nil {
		return fmt.Errorf("Failed to remove temp file, %w", err)
	}

	db_uri := fmt.Sprintf("sql://%s?dsn=%s", engine, tmp_path)
	slog.Debug(db_uri)

	db, err := NewDatabase(ctx, db_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new database for %s, %v", db_uri, err)
	}

	pt := orb.Point([]float64{-73.60033, 45.524115})

	loc := &Location{
		ID:       "1",
		Name:     "Open Da Night",
		Address:  "124 rue St. Viateur o. Montreal",
		Centroid: &pt,
	}

	err = db.AddLocation(ctx, loc)

	if err != nil {
		return fmt.Errorf("Failed to add location, %w", err)
	}

	loc, err = db.GetById(ctx, "1")

	if err != nil {
		return fmt.Errorf("Failed to retrieve location, %w", err)
	}

	// To do: GetByGeohash, etc.

	err = db.Close(ctx)

	if err != nil {
		return fmt.Errorf("Failed to close database for %s, %v", db_uri, err)
	}

	return nil
}
