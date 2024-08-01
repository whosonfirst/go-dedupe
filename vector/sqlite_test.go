package vector

import (
	"context"
	"net/url"
	"testing"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-dedupe/location"
)

func TestSQLiteDatabase(t *testing.T) {

	ctx := context.Background()

	q := url.Values{}
	q.Set("dsn", "test.db")
	q.Set("embedder-uri", "ollama://")

	u := url.URL{}
	u.Scheme = "sqlite"
	u.RawQuery = q.Encode()

	db_uri := u.String()

	db, err := NewDatabase(ctx, db_uri)

	if err != nil {
		t.Fatalf("Failed to create database for '%s', %v", db_uri, err)
	}

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

}
