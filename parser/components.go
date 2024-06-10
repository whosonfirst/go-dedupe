package parser

import (
	"fmt"

	"github.com/mmcloughlin/geohash"
	"github.com/paulmach/orb"
)

type Components struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Centroid *orb.Point        `json:"centroid"`
	Custom   map[string]string `json:"custom"`
}

func (c *Components) String() string {
	return fmt.Sprintf("[%s] %s", c.ID, c.Content())
}

func (c *Components) Content() string {
	return fmt.Sprintf("A venue named %s, located at %s", c.Name, c.Address)
}

func (c *Components) Metadata() map[string]string {

	m := make(map[string]string)

	for k, v := range c.Custom {
		m[k] = v
	}

	lon := c.Centroid[0]
	lat := c.Centroid[1]

	gh := geohash.EncodeIntWithPrecision(lat, lon, 5)
	m["geohash"] = fmt.Sprintf("%s", gh)

	return m
}
