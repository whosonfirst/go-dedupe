package database

// https://github.com/asg017/sqlite-vss/blob/main/examples/go/demo.go

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync/atomic"

	aa_sqlite "github.com/aaronland/go-sqlite"
	_ "github.com/asg017/sqlite-vss/bindings/go"
	_ "github.com/mattn/go-sqlite3"
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

	enc_e, err := db.getEmbeddings(ctx, text)

	if err != nil {
		return err
	}

	row_id := atomic.AddInt64(&db.rows, 1)

	q := fmt.Sprintf("INSERT INTO %s(rowid, embeddings) VALUES (?, ?)", sqlite_locations_vss_table)

	sum := sha256.Sum256(enc_e)
	log.Printf("%s: %d, %x\n", text, row_id, sum)

	_, err = db.database.ExecContext(ctx, q, row_id, string(enc_e))

	if err != nil {
		return err
	}

	return nil
}

func (db *SQLiteDatabase) Query(ctx context.Context, text string) ([]*QueryResult, error) {

	enc_e, err := db.getEmbeddings(ctx, text)

	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("SELECT rowid, distance FROM %s WHERE vss_search(embeddings, ?)", sqlite_locations_vss_table)

	sum := sha256.Sum256(enc_e)
	log.Printf("%s: %x\n", text, sum)

	rows, err := db.database.QueryContext(ctx, q, string(enc_e))

	if err != nil {
		log.Println("SAD")
		return nil, err
	}

	results := make([]*QueryResult, 0)

	for rows.Next() {

		var rowid int64
		var distance float32

		err = rows.Scan(&rowid, &distance)
		if err != nil {
			return nil, err
		}

		qr := &QueryResult{
			ID:         strconv.FormatInt(rowid, 10),
			Similarity: distance,
		}

		results = append(results, qr)
	}

	log.Println("COUNT", len(results), q)
	return results, nil
}

func (db *SQLiteDatabase) getEmbeddings(ctx context.Context, text string) ([]byte, error) {

	e, err := db.embedder.Embeddings(ctx, text)

	if err != nil {
		return nil, err
	}

	enc_e, err := json.Marshal(e)

	if err != nil {
		return nil, err
	}

	return enc_e, nil
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

	has_table, err := aa_sqlite.HasTableWithSQLDB(ctx, db.database, table)

	if err != nil {
		return err
	}

	if has_table {
		return nil
	}

	fname := fmt.Sprintf("sqlite_%s.schema", table)
	schema, err := FS.ReadFile(fname)

	if err != nil {
		return err
	}

	_, err = db.database.ExecContext(ctx, string(schema))

	if err != nil {
		return err
	}

	return nil
}
