package whosonfirst

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mmcloughlin/geohash"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-dedupe"
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

func (p *WhosOnFirstVenueParser) Parse(ctx context.Context, body []byte) (*parser.Components, error) {

	id, err := properties.Id(body)

	if err != nil {
		return nil, err
	}

	name, err := properties.Name(body)

	if err != nil {
		return nil, err
	}

	content := []string{
		name,
	}

	addr_rsp := gjson.GetBytes(body, "properties.addr:full")

	if addr_rsp.Exists() {

		// Something something something libpostal...

		content = append(content, addr_rsp.String())
	}

	metadata := make(map[string]string)

	country := properties.Country(body)
	metadata["country"] = country

	centroid, _, err := properties.Centroid(body)

	if err != nil {
		return nil, err
	}

	lon := centroid[0]
	lat := centroid[1]

	gh := geohash.EncodeIntWithPrecision(lat, lon, p.precision)
	metadata["geohash"] = fmt.Sprintf("%s", gh)

	c := &parser.Components{
		ID:       strconv.FormatInt(id, 10),
		Content:  strings.Join(content, " "),
		Metadata: metadata,
	}

	return c, nil
}
