package foursquare

import (
	"context"
	"fmt"
	_ "log/slog"
	"strings"

	"github.com/paulmach/orb"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/location"
)

type FoursquarePlaceParser struct {
	location.Parser
	addr_keys []string
}

func init() {
	ctx := context.Background()
	err := location.RegisterParser(ctx, "foursquare", NewFoursquarePlaceParser)

	if err != nil {
		panic(err)
	}
}

func NewFoursquarePlaceParser(ctx context.Context, uri string) (location.Parser, error) {

	addr_keys := []string{
		"address",
		"po_box",
		"post_town",
		"region",
		"admin_regin",
		"post_code",
		"country",
	}

	p := &FoursquarePlaceParser{
		addr_keys: addr_keys,
	}

	return p, nil
}

func (p *FoursquarePlaceParser) Parse(ctx context.Context, body []byte) (*location.Location, error) {

	id_rsp := gjson.GetBytes(body, "fsq_place_id")
	id := id_rsp.String()

	name_rsp := gjson.GetBytes(body, "name")
	name := name_rsp.String()

	addr_components := make([]string, 0)

	for _, k := range p.addr_keys {

		rsp := gjson.GetBytes(body, k)

		if rsp.Exists() && rsp.String() != "" {
			addr_components = append(addr_components, rsp.String())
		}
	}

	if len(addr_components) == 0 {
		return nil, dedupe.InvalidRecord(id, fmt.Errorf("Missing 'address' properties"))
	}

	// Something something something libpostal...

	addr := strings.Join(addr_components, " ")

	lat_rsp := gjson.GetBytes(body, "latitude")
	lon_rsp := gjson.GetBytes(body, "longitude")

	lat := lat_rsp.Float()
	lon := lon_rsp.Float()

	centroid := orb.Point([2]float64{lon, lat})

	c := &location.Location{
		ID:       id,
		Name:     name,
		Address:  addr,
		Centroid: &centroid,
	}

	return c, nil
}
