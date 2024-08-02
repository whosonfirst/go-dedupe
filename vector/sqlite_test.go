package vector

import (
	"context"
	_ "fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-dedupe/location"
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
			max_distance:     4,
		},
		settings{
			compression:      "matroyshka",
			expected_results: []int{1, 1, 0, 0},
			max_distance:     0.5,
		},
	}

	for _, settings := range tests {

		tmp_f, err := os.CreateTemp("", "sqlite-vec.*.db")

		if err != nil {
			t.Fatalf("Failed to create temp db, %v", err)
		}

		err = tmp_f.Close()

		if err != nil {
			t.Fatalf("Failed to close tmp db, %v", err)
		}

		tmp_name := tmp_f.Name()

		defer func() {
			slog.Debug("Remove temp db", "path", tmp_name)
			os.Remove(tmp_name)
		}()

		results := settings.expected_results

		q := url.Values{}
		q.Set("dsn", tmp_name)
		// q.Set("embedder-uri", "ollama://?model=llama3.1")
		// q.Set("dimensions", "4096")

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

		defer db.Close(ctx)

		pt := orb.Point([]float64{-73.60033, 45.524115})

		loc := &location.Location{
			ID:       "1",
			Name:     "Open Da Night",
			Address:  "124 rue St. Viateur o. Montreal",
			Centroid: &pt,
		}

		err = db.Add(ctx, loc)

		if err != nil {
			t.Fatalf("Failed to add location, %v", err)
		}

		// Add it a second time to make sure we can update
		err = db.Add(ctx, loc)

		if err != nil {
			t.Fatalf("Failed to add location, %v", err)
		}

		qr, err := db.Query(ctx, loc)

		if err != nil {
			t.Fatalf("Failed to query location, %v", err)
		}

		r1 := len(qr)
		expected1 := results[0]

		if expected1 != -1 && r1 != expected1 {
			t.Fatalf("Expected %d result(s) for query (1), but got %d", expected1, r1)
		}

		//

		loc2 := &location.Location{
			ID:       "1",
			Name:     "Open Da Night",
			Address:  "124 St. Viateur Montréal",
			Centroid: &pt,
		}

		qr2, err := db.Query(ctx, loc2)

		if err != nil {
			t.Fatalf("Failed to query location, %v", err)
		}

		r2 := len(qr2)
		expected2 := results[1]

		if expected2 != -1 && r2 != expected2 {
			t.Fatalf("Expected %d result(s) for query (1), but got %d", expected2, r2)
		}

		//

		loc3 := &location.Location{
			ID:       "1",
			Name:     "Cafe Olympico",
			Address:  "124 St. Viateur Montréal",
			Centroid: &pt,
		}

		qr3, err := db.Query(ctx, loc3)

		if err != nil {
			t.Fatalf("Failed to query location, %v", err)
		}

		r3 := len(qr3)
		expected3 := results[2]

		if expected3 != -1 && r3 != expected3 {
			t.Fatalf("Expected %d result(s) for query (1), but got %d", expected3, r3)
		}

		//

		pt4 := orb.Point([2]float64{-73.614349, 45.532726})

		loc4 := &location.Location{
			ID:       "1",
			Name:     "Cafe Italia",
			Address:  "6840 Boul Saint-Laurent",
			Centroid: &pt4,
		}

		qr4, err := db.Query(ctx, loc4)

		if err != nil {
			t.Fatalf("Failed to query location, %v", err)
		}

		r4 := len(qr4)
		expected4 := results[3]

		if expected4 != -1 && r4 != expected4 {
			t.Fatalf("Expected %d result(s) for query (1), but got %d", expected4, r4)
		}

		slog.Debug("------")
	}
}
