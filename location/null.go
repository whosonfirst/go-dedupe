package location

import (
	"context"
	"fmt"
)

type NullDatabase struct{}

func init() {
	ctx := context.Background()
	err := RegisterDatabase(ctx, "null", NewNullDatabase)

	if err != nil {
		panic(err)
	}
}

func NewNullDatabase(ctx context.Context, uri string) (Database, error) {
	db := &NullDatabase{}
	return db, nil
}

func (db *NullDatabase) AddLocation(ctx context.Context, loc *Location) error {
	return nil
}

func (db *NullDatabase) GetById(ctx context.Context, id string) (*Location, error) {
	return nil, fmt.Errorf("Not found")
}

func (db *NullDatabase) GetGeohashes(ctx context.Context, cb GetGeohashesCallback) error {
	return nil
}

func (db *NullDatabase) GetWithGeohash(ctx context.Context, geohash string, cb GetWithGeohashCallback) error {
	return nil
}

func (db *NullDatabase) Close(ctx context.Context) error {
	return nil
}
