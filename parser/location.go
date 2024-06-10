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
	// Something something libpostal c.Address...
	return fmt.Sprintf("A venue named %s, located at %s", c.Name, c.Address)
}

func (c *Location) Metadata() map[string]string {

	m := make(map[string]string)

	for k, v := range c.Custom {
		m[k] = v
	}

	lon := c.Centroid[0]
	lat := c.Centroid[1]

	gh := geohash.EncodeIntWithPrecision(lat, lon, 5)
	m["geohash"] = fmt.Sprintf("%s", gh)

	// Something something libpostal parse c.Address... and add components as metadata?
	return m
}
