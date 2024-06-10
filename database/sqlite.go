package database

// https://github.com/asg017/sqlite-vss/blob/main/examples/go/demo.go

import (
	"context"
	_ "embed"

	"github.com/whosonfirst/go-dedupe"
)

//go:embed sqlite.schema
var sqlite_schema string

type SQLiteDatabase struct {
}

func init() {
	ctx := context.Background()
	err := RegisterDatabase(ctx, "sqlite", NewSQLiteDatabase)

	if err != nil {
		panic(err)
	}
}

func NewSQLiteDatabase(ctx context.Context, uri string) (Database, error) {
	return nil, dedupe.NotImplemented()
}

func (db *SQLiteDatabase) Add(ctx context.Context, id string, text string, metadata map[string]string) error {
	return dedupe.NotImplemented()
}

func (db *SQLiteDatabase) Query(ctx context.Context, text string) ([]*QueryResult, error) {
	return nil, dedupe.NotImplemented()
}
