package location

import (
	"context"
	"fmt"
	_ "log/slog"
	"slices"

	"github.com/mmcloughlin/geohash"
	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-dedupe/embeddings"
)

var reserved_metadata_keys = []string{
	"geohash",
}

// Location defines a common format for locations for the purposes of deduplication
type Location struct {
	// The unique ID for the location which is expected to take the form of "{SOURCE_PREFIX}:id={UNIQUE ID}"
	ID string `json:"id"`
	// The name of the location
	Name string `json:"name"`
	// The complete address of the location
	Address string `json:"address"`
	// The principal centroid for the location
	Centroid *orb.Point `json:"centroid"`
	// An arbitrary dictionary of custom metadata properties for the location
	Custom map[string]string `json:"custom,omitempty"`
}

// String returns the locations name and address as a comma-separated string.
func (loc *Location) String() string {
	return fmt.Sprintf("%s, %s", loc.Name, loc.Address)
}

// Metadata returns the union of automatically derived metadata properties (geohash) and any custom metadata properties.
func (loc *Location) Metadata() map[string]string {

	m := make(map[string]string)

	for k, v := range loc.Custom {

		/*
			if IsReservedMetadataKey(k) {
				slog.Warn("Location metadata contains reserved key, skipping", "location", loc.ID, "key", k)
				continue
			}
		*/

		m[k] = v
	}

	if loc.Centroid != nil {
		m["geohash"] = loc.Geohash()
	}

	// Something something libpostal parse c.Address... and add components as metadata?
	return m
}

// Geohash returns the geohash with a precision of 5 for the location.
func (loc *Location) Geohash() string {
	lon := loc.Centroid[0]
	lat := loc.Centroid[1]
	return geohash.EncodeWithPrecision(lat, lon, 5)
}

func (loc *Location) Embeddings32(ctx context.Context, embedder embeddings.Embedder) ([]float32, error) {

	text := fmt.Sprintf("%s, %s", loc.Name, loc.Address)

	return embedder.Embeddings32(ctx, text)
}

// ReservedMetadataKeys returns the list of reserved metadata keys.
func ReservedMetadataKeys() []string {
	return reserved_metadata_keys
}

// IsReservedMetadataKeys returns a boolean indicating whether 'k' is reserved.
func IsReservedMetadataKey(k string) bool {
	return slices.Contains(reserved_metadata_keys, k)
}
