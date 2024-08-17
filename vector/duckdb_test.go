package vector

import (
	"context"
	"log/slog"
	"net/url"
	"testing"
)

func TestDuckDBDatabase(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx := context.Background()

	q := url.Values{}

	q.Set("embedder-uri", "ollama://?model=mxbai-embed-large")
	q.Set("dimensions", "1024")
	q.Set("max-results", "10")
	// q.Set("max-distance", strconv.FormatFloat(float64(settings.max_distance), 'g', -1, 32))

	u := url.URL{}
	u.Scheme = "duckdb"
	u.RawQuery = q.Encode()

	db_uri := u.String()
	slog.Debug(db_uri)

	db, err := NewDatabase(ctx, db_uri)

	if err != nil {
		t.Fatalf("Failed to create duckdb database, %v", err)
	}

	err = db.Close(ctx)

	if err != nil {
		t.Fatalf("Failed to close duckdb database, %v", err)
	}
}
