package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
)

type Table struct {
	Name   string
	Schema string
}

type ConfigureDatabaseOptions struct {
	CreateTablesIfNecessary bool
	Tables                  []*Table
	Pragma                  []string
}

func DefaultConfigureDatabaseOptions() *ConfigureDatabaseOptions {

	opts := &ConfigureDatabaseOptions{
		Pragma: DefaultPragma(),
	}

	return opts
}

func DefaultPragma() []string {

	pragma := []string{
		"PRAGMA JOURNAL_MODE=OFF",
		"PRAGMA SYNCHRONOUS=OFF",
		// "PRAGMA LOCKING_MODE=EXCLUSIVE",
		// https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html
		"PRAGMA PAGE_SIZE=4096",
		"PRAGMA CACHE_SIZE=1000000",
	}

	return pragma
}

func ConfigureDatabase(ctx context.Context, db *sql.DB, opts *ConfigureDatabaseOptions) error {

	if opts.CreateTablesIfNecessary {

		table_names := make([]string, 0)

		sql := "SELECT name FROM sqlite_master WHERE type='table'"

		rows, err := db.QueryContext(ctx, sql)

		if err != nil {
			return fmt.Errorf("Failed to query sqlite_master, %w", err)
		}

		defer rows.Close()

		for rows.Next() {

			var name string
			err := rows.Scan(&name)

			if err != nil {
				return fmt.Errorf("Failed scan table name, %w", err)
			}

			table_names = append(table_names, name)
		}

		for _, t := range opts.Tables {

			if slices.Contains(table_names, t.Name) {
				continue
			}

			slog.Debug("Create table", "name", t.Name, "schema", t.Schema)

			_, err := db.ExecContext(ctx, t.Schema)

			if err != nil {
				return fmt.Errorf("Failed to create %s table, %w", t.Name, err)
			}
		}
	}

	for _, p := range opts.Pragma {

		_, err := db.ExecContext(ctx, p)

		if err != nil {
			return fmt.Errorf("Failed to set pragma '%s', %w", p, err)
		}
	}

	return nil
}
