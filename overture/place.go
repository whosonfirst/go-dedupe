package overture

import (
	"context"
	"fmt"
	_ "log/slog"
	"strings"

	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/parser"
)

type OverturePlaceParser struct {
	parser.Parser
	precision uint
	addr_keys []string
}

func init() {
	ctx := context.Background()
	err := parser.RegisterParser(ctx, "overtureplaces", NewOverturePlaceParser)

	if err != nil {
		panic(err)
	}
}

func NewOverturePlaceParser(ctx context.Context, uri string) (parser.Parser, error) {

	addr_keys := []string{
		"freeform",
		"locality",
		"region",
		"country",
	}

	p := &OverturePlaceParser{
		precision: dedupe.DEFAULT_GEOHASH_PRECISION,
		addr_keys: addr_keys,
	}

	return p, nil
}

func (p *OverturePlaceParser) Parse(ctx context.Context, body []byte) (*parser.Components, error) {

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

		addr_components := make([]string, 0)

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

	c := &parser.Components{
		ID:       c_id,
		Name:     name,
		Address:  addr,
		Centroid: &centroid,
	}

	return c, nil
}
