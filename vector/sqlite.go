package vector

import (
	"database/sql"
	"fmt"
	"net/url"

	"context"
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whosonfirst/go-dedupe/embeddings"
	"github.com/whosonfirst/go-dedupe/location"
)

type SQLiteDatabase struct {
	vec_db   *sql.DB
	embedder embeddings.Embedder
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

	vec_db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	embedder_uri := q.Get("embedder-uri")

	embdr, err := embeddings.NewEmbedder(ctx, embedder_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new embedder, %w", err)
	}

	db := &SQLiteDatabase{
		vec_db:   vec_db,
		embedder: embdr,
	}

	return db, nil
}

func (db *SQLiteDatabase) Add(ctx context.Context, loc *location.Location) error {

	id := loc.ID

	q, err := loc.Embeddings32(ctx, db.embedder)

	if err != nil {
		return fmt.Errorf("Failed to derive query for location, %w", err)
	}

	v, err := sqlite_vec.SerializeFloat32(q)

	if err != nil {
		return fmt.Errorf("Failed to serialize floats for ID %d, %w", id, err)
	}

	_, err = db.vec_db.ExecContext(ctx, "INSERT INTO vec_items(rowid, embedding) VALUES (?, ?)", id, v)

	if err != nil {
		return fmt.Errorf("Failed to insert row for ID %d, %w", id, err)
	}

	return nil
}

func (db *SQLiteDatabase) Query(ctx context.Context, loc *location.Location) ([]*QueryResult, error) {

	results := make([]*QueryResult, 0)

	q, err := loc.Embeddings32(ctx, db.embedder)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive query for location, %w", err)
	}

	query, err := sqlite_vec.SerializeFloat32(q)

	if err != nil {
		return nil, fmt.Errorf("Failed to serialize query, %w", err)
	}

	rows, err := db.vec_db.QueryContext(ctx, `SELECT rowid, distance FROM vec_items WHERE embedding MATCH ? ORDER BY distance`, query)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute query, %w", err)
	}

	for rows.Next() {

		var rowid int64
		var distance float64

		err = rows.Scan(&rowid, &distance)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan row, %w", err)
		}

		fmt.Printf("rowid=%d, distance=%f\n", rowid, distance)
	}

	return results, nil
}

func (db *SQLiteDatabase) Flush(ctx context.Context) error {
	return nil
}

func (db *SQLiteDatabase) Close(ctx context.Context) error {
	return db.vec_db.Close()
}
