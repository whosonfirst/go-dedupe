package vector

import (
	"context"
	"log/slog"
	"net/url"
	"strconv"
	"testing"
)

func TestSQLiteDatabase(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx := context.Background()

	type settings struct {
		compression      string
		expected_results []int
		max_distance     float32
	}

	tests := []settings{
		settings{
			compression:      "none",
			expected_results: []int{1, 1, 0, 0},
			max_distance:     4,
		},
		settings{
			compression:      "quantize",
			expected_results: []int{1, 0, 0, 0},
			max_distance:     1,
		},
		settings{
			compression:      "matroyshka",
			expected_results: []int{1, 1, 0, 0},
			max_distance:     0.5,
		},
	}

	for _, settings := range tests {

		expected_results := settings.expected_results

		q := url.Values{}
		q.Set("dsn", "{tmp}.db")

		q.Set("embedder-uri", "ollama://?model=mxbai-embed-large")
		q.Set("dimensions", "1024")
		q.Set("max-results", "10")
		q.Set("max-distance", strconv.FormatFloat(float64(settings.max_distance), 'g', -1, 32))
		q.Set("compression", settings.compression)

		u := url.URL{}
		u.Scheme = "sqlite"
		u.RawQuery = q.Encode()

		db_uri := u.String()

		slog.Debug("Create new vector database", "uri", db_uri)

		db, err := NewDatabase(ctx, db_uri)

		if err != nil {
			t.Fatalf("Failed to create database for '%s', %v", db_uri, err)
		}

		err = testDatabaseWithLocations(ctx, db, expected_results)

		if err != nil {
			t.Fatal(err)
		}

		err = db.Close(ctx)

		if err != nil {
			t.Fatalf("Failed to close database, %v", err)
		}

	}
}
