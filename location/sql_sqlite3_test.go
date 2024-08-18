//go:build sqlite3

package location

import (
	"context"
	"log/slog"
	"testing"
)

func TestSQLite3Database(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx := context.Background()

	err := testSQLDatabaseEngine(ctx, "sqlite3")

	if err != nil {
		t.Fatal(err)
	}
}
