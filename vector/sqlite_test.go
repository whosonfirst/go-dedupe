package vector

import (
	"context"
	_ "fmt"
	"log/slog"
	"net/url"
	"os"
	"testing"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-dedupe/location"
)

func TestSQLiteDatabase(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx := context.Background()

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

	q := url.Values{}
	q.Set("dsn", tmp_name)
	// q.Set("embedder-uri", "ollama://?model=llama3.1")
	// q.Set("dimensions", "4096")

	q.Set("embedder-uri", "ollama://?model=mxbai-embed-large")
	q.Set("dimensions", "1024")
	q.Set("max-results", "10")
	q.Set("max-distance", "4")

	u := url.URL{}
	u.Scheme = "sqlite"
	u.RawQuery = q.Encode()

	db_uri := u.String()

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

	r, err := db.Query(ctx, loc)

	if err != nil {
		t.Fatalf("Failed to query location, %v", err)
	}

	if len(r) != 1 {
		t.Fatalf("Expected one result for query (1), but got %d", len(r))
	}

	loc2 := &location.Location{
		ID:       "1",
		Name:     "Open Da Night",
		Address:  "124 St. Viateur Montréal",
		Centroid: &pt,
	}

	r2, err := db.Query(ctx, loc2)

	if err != nil {
		t.Fatalf("Failed to query location, %v", err)
	}

	if len(r2) != 1 {
		t.Fatalf("Expected one result for query (2), but got %d", len(r2))
	}

	loc3 := &location.Location{
		ID:       "1",
		Name:     "Cafe Olympico",
		Address:  "124 St. Viateur Montréal",
		Centroid: &pt,
	}

	r3, err := db.Query(ctx, loc3)

	if err != nil {
		t.Fatalf("Failed to query location, %v", err)
	}

	if len(r3) != 0 {
		t.Fatalf("Expected zero results for query (3), but got %d", len(r3))
	}

	pt4 := orb.Point([2]float64{-73.614349, 45.532726})

	loc4 := &location.Location{
		ID:       "1",
		Name:     "Cafe Italia",
		Address:  "6840 Boul Saint-Laurent",
		Centroid: &pt4,
	}

	r4, err := db.Query(ctx, loc4)

	if err != nil {
		t.Fatalf("Failed to query location, %v", err)
	}

	if len(r4) != 0 {
		t.Fatalf("Expected zero results for query (4), but got %d", len(r4))
	}

}
