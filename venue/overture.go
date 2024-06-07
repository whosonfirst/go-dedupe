package venue

import (
	"context"
	"fmt"
	"strings"
	_ "log/slog"
	
	"github.com/mmcloughlin/geohash"
	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-dedupe/parser"
)

type OvertureVenueParser struct {
	parser.Parser
	precision uint
	addr_keys []string
}

func init() {
	ctx := context.Background()
	err := parser.RegisterParser(ctx, "overtureplaces", NewOvertureVenueParser)

	if err != nil {
		panic(err)
	}
}

func NewOvertureVenueParser(ctx context.Context, uri string) (parser.Parser, error) {

	addr_keys := []string{
		"freeform",
		"locality",
		"region",
		"country",
	}

	p := &OvertureVenueParser{
		precision: DEFAULT_GEOHASH_PRECISION,
		addr_keys: addr_keys,
	}

	return p, nil
}

func (p *OvertureVenueParser) Parse(ctx context.Context, body []byte) (*parser.Components, error) {

	id_rsp := gjson.GetBytes(body, "properties.id")

	if !id_rsp.Exists() {
		return nil, fmt.Errorf("Missing 'id' property")
	}

	ovtr_id := id_rsp.String()
	
	name_rsp := gjson.GetBytes(body, "properties.names.primary")

	content := []string{
		name_rsp.String(),
	}

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

		// Something something something libpostal...

		content = append(content, strings.Join(addr_components, " "))
	}

	geom_rsp := gjson.GetBytes(body, "geometry")

	geom, err := geojson.UnmarshalGeometry([]byte(geom_rsp.String()))

	if err != nil {
		return nil, err
	}

	f := geojson.NewFeature(geom.Geometry())
	centroid := f.Point()

	metadata := make(map[string]string)

	lon := centroid[0]
	lat := centroid[1]

	gh := geohash.EncodeWithPrecision(lat, lon, p.precision)
	metadata["geohash"] = gh

	c := &parser.Components{
		ID:       ovtr_id,
		Content:  strings.Join(content, " "),
		Metadata: metadata,
	}

	return c, nil
}
