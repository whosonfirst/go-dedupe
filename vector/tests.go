package vector

import (
	"context"
	"fmt"

	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-dedupe/location"
)

func testDatabaseWithLocations(ctx context.Context, db Database, results []int) error {

	//

	pt := orb.Point([]float64{-73.60033, 45.524115})

	loc := &location.Location{
		ID:       "1",
		Name:     "Open Da Night",
		Address:  "124 rue St. Viateur o. Montreal",
		Centroid: &pt,
	}

	err := db.Add(ctx, loc)

	if err != nil {
		return fmt.Errorf("Failed to add location, %w", err)
	}

	// Add it a second time to make sure we can update
	err = db.Add(ctx, loc)

	if err != nil {
		return fmt.Errorf("Failed to add location twice, %w", err)
	}

	qr, err := db.Query(ctx, loc)

	if err != nil {
		return fmt.Errorf("Failed to query location, %w", err)
	}

	r1 := len(qr)
	expected1 := results[0]

	if expected1 != -1 && r1 != expected1 {
		return fmt.Errorf("Expected %d result(s) for query (1), but got %d", expected1, r1)
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
		return fmt.Errorf("Failed to query location, %w", err)
	}

	r2 := len(qr2)
	expected2 := results[1]

	if expected2 != -1 && r2 != expected2 {
		return fmt.Errorf("Expected %d result(s) for query (1), but got %d", expected2, r2)
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
		return fmt.Errorf("Failed to query location, %w", err)
	}

	r3 := len(qr3)
	expected3 := results[2]

	if expected3 != -1 && r3 != expected3 {
		return fmt.Errorf("Expected %d result(s) for query (1), but got %d", expected3, r3)
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
		return fmt.Errorf("Failed to query location, %w", err)
	}

	r4 := len(qr4)
	expected4 := results[3]

	if expected4 != -1 && r4 != expected4 {
		return fmt.Errorf("Expected %d result(s) for query (1), but got %d", expected4, r4)
	}

	return nil
}
