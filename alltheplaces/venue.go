package alltheplaces

import (
	"context"
	"fmt"
	_ "log"
	"log/slog"
	"strings"

	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-dedupe/parser"
)

type AllThePlacesVenueParser struct {
	parser.Parser
	precision uint
	addr_keys []string
}

func init() {
	ctx := context.Background()
	err := parser.RegisterParser(ctx, "alltheplaces", NewAllThePlacesVenueParser)

	if err != nil {
		panic(err)
	}
}

func NewAllThePlacesVenueParser(ctx context.Context, uri string) (parser.Parser, error) {

	addr_keys := []string{
		"addr:street_address",
		"addr:city",
		"addr:state",
		"addr:country",
	}

	p := &AllThePlacesVenueParser{
		precision: dedupe.DEFAULT_GEOHASH_PRECISION,
		addr_keys: addr_keys,
	}

	return p, nil
}

func (p *AllThePlacesVenueParser) Parse(ctx context.Context, body []byte) (*location.Location, error) {

	id_rsp := gjson.GetBytes(body, "id")

	if !id_rsp.Exists() {
		return nil, dedupe.InvalidRecord("#", fmt.Errorf("Missing 'id' property"))
	}

	id := id_rsp.String()

	name_rsp := gjson.GetBytes(body, "properties.name")

	if !name_rsp.Exists() {
		return nil, dedupe.InvalidRecord(id, fmt.Errorf("Missing 'name' property"))
	}

	name := name_rsp.String()

	addr_components := make([]string, 0)

	for _, k := range p.addr_keys {

		path := fmt.Sprintf("properties.%s", k)
		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() && rsp.String() != "" {
			addr_components = append(addr_components, rsp.String())
		}
	}

	if len(addr_components) == 0 {
		return nil, dedupe.InvalidRecord(id, fmt.Errorf("Missing 'address' properties"))
	}

	// Something something something libpostal...

	addr := strings.Join(addr_components, " ")

	geom_rsp := gjson.GetBytes(body, "geometry")

	if !geom_rsp.Exists() || geom_rsp.String() == "" {
		slog.Warn("Record is missing geometry", "id", id)
		return nil, dedupe.InvalidRecord(id, nil)
	}

	geom, err := geojson.UnmarshalGeometry([]byte(geom_rsp.String()))

	if err != nil {
		return nil, err
	}

	f := geojson.NewFeature(geom.Geometry())
	centroid := f.Point()

	c_id := dedupe.AllThePlacesId(id)

	c := &location.Location{
		ID:       c_id,
		Name:     name,
		Address:  addr,
		Centroid: &centroid,
	}

	return c, nil
}
