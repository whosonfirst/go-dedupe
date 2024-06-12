package location

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/mmcloughlin/geohash"
	"github.com/paulmach/orb"
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

func (loc *Location) String() string {
	return fmt.Sprintf("%s, %s", loc.Name, loc.Address)
}

func (loc *Location) Metadata() map[string]string {

	m := make(map[string]string)

	for k, v := range loc.Custom {

		if IsReservedMetadataKey(k) {
			slog.Warn("Location metadata contains reserved key, skipping", "location", loc.ID, "key", k)
			continue
		}

		m[k] = v
	}

	m["geohash"] = loc.Geohash()

	// Something something libpostal parse c.Address... and add components as metadata?
	return m
}

func (loc *Location) Geohash() string {
	lon := loc.Centroid[0]
	lat := loc.Centroid[1]
	return geohash.EncodeWithPrecision(lat, lon, 5)
}

func ReservedMetadataKeys() []string {
	return reserved_metadata_keys
}

func IsReservedMetadataKey(k string) bool {
	return slices.Contains(reserved_metadata_keys, k)
}
