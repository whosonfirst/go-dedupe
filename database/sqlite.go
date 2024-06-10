package database

// https://github.com/asg017/sqlite-vss/blob/main/examples/go/demo.go

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync/atomic"

	aa_sqlite "github.com/aaronland/go-sqlite"
	_ "github.com/asg017/sqlite-vss/bindings/go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whosonfirst/go-dedupe"
	"github.com/whosonfirst/go-dedupe/embeddings"
)

// #cgo linux,amd64 LDFLAGS: -Wl,-undefined,dynamic_lookup -lstdc++
// #cgo darwin,amd64 LDFLAGS: -Wl,-undefined,dynamic_lookup -lomp
// #cgo darwin,arm64 LDFLAGS: -Wl,-undefined,dynamic_lookup -lomp
import "C"

//go:embed sqlite_*.schema
var FS embed.FS

const sqlite_locations_table string = "locations"
const sqlite_locations_vss_table string = "locations_vss"

type SQLiteDatabase struct {
	database *sql.DB
	embedder embeddings.Embedder
	rows     int64
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

	log.Println(engine, dsn)

	sql_db, err := sql.Open(engine, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	emb_uri := "ollama://?model=mxbai-embed-large"
	emb, err := embeddings.NewEmbedder(ctx, emb_uri)

	if err != nil {
		return nil, err
	}

	db := &SQLiteDatabase{
		database: sql_db,
		embedder: emb,
	}

	err = db.setupTables(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to set up tables, %w", err)
	}

	return db, nil
}

func (db *SQLiteDatabase) Add(ctx context.Context, id string, text string, metadata map[string]string) error {

	e, err := db.embedder.Embeddings(ctx, text)

	if err != nil {
		return err
	}

	enc_e, err := json.Marshal(e)

	if err != nil {
		return err
	}

	row_id := atomic.AddInt64(&db.rows, 1)
	log.Println("ROW ID", row_id)

	q := fmt.Sprintf("INSERT INTO %s(rowid, embeddings) VALUES (?, ?)", sqlite_locations_vss_table)

	_, err = db.database.ExecContext(ctx, q, row_id, string(enc_e))

	if err != nil {
		return err
	}

	return nil
}

func (db *SQLiteDatabase) Query(ctx context.Context, text string) ([]*QueryResult, error) {
	return nil, dedupe.NotImplemented()
}

func (db *SQLiteDatabase) setupTables(ctx context.Context) error {

	tables := []string{
		sqlite_locations_table,
		sqlite_locations_vss_table,
	}

	for _, t := range tables {

		err := db.setupTable(ctx, t)

		if err != nil {
			return err
		}
	}

	q := fmt.Sprintf("SELECT COUNT(rowid) AS count FROM %s", sqlite_locations_vss_table)
	row := db.database.QueryRowContext(ctx, q)

	var max int64

	err := row.Scan(&max)

	if err != nil {
		return err
	}

	atomic.StoreInt64(&db.rows, max)

	return nil
}

func (db *SQLiteDatabase) setupTable(ctx context.Context, table string) error {

	log.Println("Table", table)

	has_table, err := aa_sqlite.HasTableWithSQLDB(ctx, db.database, table)

	if err != nil {
		return err
	}

	if has_table {
		log.Println("exists", table)
		return nil
	}

	fname := fmt.Sprintf("sqlite_%s.schema", table)
	schema, err := FS.ReadFile(fname)

	if err != nil {
		return err
	}

	log.Println("WTF", table, string(schema))

	_, err = db.database.ExecContext(ctx, string(schema))

	if err != nil {
		return err
	}

	log.Println("POO", table)
	return nil
}
