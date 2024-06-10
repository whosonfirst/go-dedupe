package parser

import (
	"fmt"

	"github.com/mmcloughlin/geohash"
	"github.com/paulmach/orb"
)

type Location struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Centroid *orb.Point        `json:"centroid"`
	Custom   map[string]string `json:"custom"`
}

func (c *Location) String() string {
	return fmt.Sprintf("[%s] %s", c.ID, c.Content())
}

func (c *Location) Content() string {
	// Something something libpostal c.Address... or maybe just rely on metadata for
	// structured data but "metadata" seems to be specific to philippgille/chromem-go
	// so maybe not?

	return fmt.Sprintf("A venue named %s, contained by the geohash %s, located at %s", c.Name, c.Geohash(), c.Address)
}

func (c *Location) Metadata() map[string]string {

	m := make(map[string]string)

	for k, v := range c.Custom {
		m[k] = v
	}

	m["geohash"] = c.Geohash()

	// Something something libpostal parse c.Address... and add components as metadata?
	return m
}

func (c *Location) Geohash() string {
	lon := c.Centroid[0]
	lat := c.Centroid[1]
	return geohash.EncodeWithPrecision(lat, lon, 5)
}
