package database

import (
	"context"
	"database/sql"
	"fmt"
)

func DefaultSQLitePragma() []string {

	pragma := []string{
		"PRAGMA JOURNAL_MODE=OFF",
		"PRAGMA SYNCHRONOUS=OFF",
		// https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html
		"PRAGMA PAGE_SIZE=4096",
		"PRAGMA CACHE_SIZE=1000000",
	}

	return pragma
}

func ConfigureSQLitePragma(ctx context.Context, db *sql.DB, pragma []string) error {

	for _, p := range pragma {

		_, err := db.ExecContext(ctx, p)

		if err != nil {
			return fmt.Errorf("Failed to set pragma '%s', %w", p, err)
		}
	}

	return nil
}
