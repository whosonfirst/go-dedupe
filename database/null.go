package database

import (
	"context"
	"github.com/whosonfirst/go-dedupe/location"
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

func (db *NullDatabase) Add(ctx context.Context, loc *location.Location) error {
	return nil
}

func (db *NullDatabase) Query(ctx context.Context, text string, metadata map[string]string) ([]*QueryResult, error) {
	results := make([]*QueryResult, 0)
	return results, nil
}

func (db *NullDatabase) Flush(ctx context.Context) error {
	return nil
}
