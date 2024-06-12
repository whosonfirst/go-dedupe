package whosonfirst

import (
	"context"
	"strconv"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-dedupe/parser"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

type WhosOnFirstVenueParser struct {
	parser.Parser
	precision uint
}

func init() {
	ctx := context.Background()
	err := parser.RegisterParser(ctx, "whosonfirstvenues", NewWhosOnFirstVenueParser)

	if err != nil {
		panic(err)
	}
}

func NewWhosOnFirstVenueParser(ctx context.Context, uri string) (parser.Parser, error) {

	p := &WhosOnFirstVenueParser{
		precision: dedupe.DEFAULT_GEOHASH_PRECISION,
	}

	return p, nil
}

func (p *WhosOnFirstVenueParser) Parse(ctx context.Context, body []byte) (*location.Location, error) {

	id, err := properties.Id(body)

	if err != nil {
		return nil, err
	}

	name, err := properties.Name(body)

	if err != nil {
		return nil, err
	}

	addr_rsp := gjson.GetBytes(body, "properties.addr:full")

	metadata := make(map[string]string)

	country := properties.Country(body)
	metadata["country"] = country

	centroid, _, err := properties.Centroid(body)

	if err != nil {
		return nil, err
	}

	str_id := strconv.FormatInt(id, 10)
	c_id := dedupe.WhosOnFirstId(str_id)

	c := &location.Location{
		ID:       c_id,
		Name:     name,
		Address:  addr_rsp.String(),
		Centroid: centroid,
	}

	return c, nil
}
