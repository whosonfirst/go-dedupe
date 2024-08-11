package overture

import (
	"context"
	"fmt"
	_ "log/slog"
	"strings"

	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/location"
)

type OverturePlaceParser struct {
	location.Parser
	addr_keys []string
}

func init() {
	ctx := context.Background()
	err := location.RegisterParser(ctx, "overtureplaces", NewOverturePlaceParser)

	if err != nil {
		panic(err)
	}
}

func NewOverturePlaceParser(ctx context.Context, uri string) (location.Parser, error) {

	addr_keys := []string{
		"freeform",
		"locality",
		"region",
		"country",
	}

	p := &OverturePlaceParser{
		addr_keys: addr_keys,
	}

	return p, nil
}

func (p *OverturePlaceParser) Parse(ctx context.Context, body []byte) (*location.Location, error) {

	id_rsp := gjson.GetBytes(body, "properties.id")

	if !id_rsp.Exists() {
		return nil, dedupe.InvalidRecord("#", fmt.Errorf("Missing 'id' property"))
	}

	id := id_rsp.String()

	name_rsp := gjson.GetBytes(body, "properties.names.primary")

	if !name_rsp.Exists() {
		return nil, dedupe.InvalidRecord(id, fmt.Errorf("Missing 'name' property"))
	}

	name := name_rsp.String()

	addr_components := make([]string, 0)

	addrs_rsp := gjson.GetBytes(body, "properties.addresses")

	for _, rsp := range addrs_rsp.Array() {

		addr := make(map[string]string)

		for k, v := range rsp.Map() {
			addr[k] = v.String()
		}

		for _, k := range p.addr_keys {

			v, exists := addr[k]

			if exists && v != "" {
				addr_components = append(addr_components, v)
			}
		}
	}

	if len(addr_components) == 0 {
		return nil, dedupe.InvalidRecord(id, fmt.Errorf("Missing 'address' properties"))
	}

	// Something something something libpostal...

	addr := strings.Join(addr_components, " ")

	geom_rsp := gjson.GetBytes(body, "geometry")

	geom, err := geojson.UnmarshalGeometry([]byte(geom_rsp.String()))

	if err != nil {
		return nil, err
	}

	f := geojson.NewFeature(geom.Geometry())
	centroid := f.Point()

	c_id := dedupe.OvertureId(id)

	c := &location.Location{
		ID:       c_id,
		Name:     name,
		Address:  addr,
		Centroid: &centroid,
	}

	return c, nil
}
