package location

import (
	"context"
	"fmt"
	"iter"
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

func (db *NullDatabase) GetGeohashes(ctx context.Context) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		return
	}
}

func (db *NullDatabase) GetWithGeohash(ctx context.Context, geohash string) iter.Seq2[*Location, error] {
	return func(yield func(*Location, error) bool) {
		return
	}
}

func (db *NullDatabase) Close(ctx context.Context) error {
	return nil
}
