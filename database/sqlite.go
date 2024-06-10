package database

// https://github.com/asg017/sqlite-vss/blob/main/examples/go/demo.go

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/url"

	_ "github.com/asg017/sqlite-vss/bindings/go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whosonfirst/go-dedupe"
)

// #cgo linux,amd64 LDFLAGS: -Wl,-undefined,dynamic_lookup -lstdc++
// #cgo darwin,amd64 LDFLAGS: -Wl,-undefined,dynamic_lookup -lomp
// #cgo darwin,arm64 LDFLAGS: -Wl,-undefined,dynamic_lookup -lomp
import "C"

//go:embed sqlite.schema
var sqlite_schema string

type SQLiteDatabase struct {
	database *sql.DB
}

func init() {
	ctx := context.Background()
	err := RegisterDatabase(ctx, "sqlite", NewSQLiteDatabase)

	if err != nil {
		panic(err)
	}
}

func NewSQLiteDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()
	dsn := q.Get("dsn")

	engine := "sqlite3"

	sql_db, err := sql.Open(engine, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	db := &SQLiteDatabase{
		database: sql_db,
	}

	return db, nil
}

func (db *SQLiteDatabase) Add(ctx context.Context, id string, text string, metadata map[string]string) error {
	return dedupe.NotImplemented()
}

func (db *SQLiteDatabase) Query(ctx context.Context, text string) ([]*QueryResult, error) {
	return nil, dedupe.NotImplemented()
}
