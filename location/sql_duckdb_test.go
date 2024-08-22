//go:build duckdb

package location

import (
	"context"
	"log/slog"
	"testing"
)

func TestDuckDBDatabase(t *testing.T) {

	ctx := context.Background()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	/*

		=== RUN   TestDuckDBDatabase
		    sql_duckdb_test.go:15: Failed to create new database for sql://duckdb?dsn=/var/folders/_k/h7ndzcyx3dq027gsrg1q45xm0000gn/T/1788184000-duckdb.db, Failed to open database connection, database/sql/driver: could not open database: duckdb error: IO Error: The file "/var/folders/_k/h7ndzcyx3dq027gsrg1q45xm0000gn/T/1788184000-duckdb.db" exists, but it is not a valid DuckDB database file!

	*/

	err := testSQLDatabaseEngine(ctx, "duckdb")

	if err != nil {
		t.Fatal(err)
	}
}
