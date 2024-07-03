package location

import (
	"context"
)

type GetWithGeohashCallback func(context.Context, *Location) error

type Database interface {
	AddLocation(context.Context, *Location) error
	GetById(context.Context, string) (*Location, error)
	GetWithGeohash(context.Context, string, GetWithGeohashCallback) error
}
