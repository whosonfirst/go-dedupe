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

	has_table, err := hasTable(ctx, vec_db, "vec_items")

	if err != nil {
		return nil, fmt.Errorf("Failed to determine if vec_items table exists, %w", err)
	}

	if !has_table {

		vectors := 768

		q := fmt.Sprintf("CREATE VIRTUAL TABLE vec_items USING vec0(embedding float[%d])", vectors)

		_, err := vec_db.Exec(q)

		if err != nil {
			return nil, fmt.Errorf("Failed to create vec_items table, %w", err)
		}
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

	v, err := db.embeddings(ctx, loc)

	if err != nil {
		return fmt.Errorf("Failed to serialize floats for ID %s, %w", id, err)
	}

	_, err = db.vec_db.ExecContext(ctx, "INSERT INTO vec_items(rowid, embedding) VALUES (?, ?)", id, v)

	if err != nil {
		return fmt.Errorf("Failed to insert row for ID %s, %w", id, err)
	}

	return nil
}

func (db *SQLiteDatabase) Query(ctx context.Context, loc *location.Location) ([]*QueryResult, error) {

	results := make([]*QueryResult, 0)

	query, err := db.embeddings(ctx, loc)

	if err != nil {
		return nil, fmt.Errorf("Failed to serialize query, %w", err)
	}

	rows, err := db.vec_db.QueryContext(ctx, `SELECT rowid, distance FROM vec_items WHERE embedding MATCH ? ORDER BY distance`, query)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute query, %w", err)
	}

	for rows.Next() {

		var id string
		var distance float64

		err = rows.Scan(&id, &distance)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan row, %w", err)
		}

		r := &QueryResult{
			ID:         id,
			Similarity: float32(distance),
		}

		results = append(results, r)
		// fmt.Printf("rowid=%d, distance=%f\n", rowid, distance)
	}

	return results, nil
}

func (db *SQLiteDatabase) Flush(ctx context.Context) error {
	return nil
}

func (db *SQLiteDatabase) Close(ctx context.Context) error {
	return db.vec_db.Close()
}

func (db *SQLiteDatabase) embeddings(ctx context.Context, loc *location.Location) ([]byte, error) {

	q, err := loc.Embeddings32(ctx, db.embedder)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive query for location, %w", err)
	}

	query, err := sqlite_vec.SerializeFloat32(q)

	if err != nil {
		return nil, fmt.Errorf("Failed to serialize query, %w", err)
	}

	return query, nil
}

func hasTable(ctx context.Context, db *sql.DB, table_name string) (bool, error) {

	has_table := false

	sql := "SELECT name FROM sqlite_master WHERE type='table'"

	rows, err := db.QueryContext(ctx, sql)

	if err != nil {
		return false, fmt.Errorf("Failed to query sqlite_master, %w", err)
	}

	defer rows.Close()

	for rows.Next() {

		var name string
		err := rows.Scan(&name)

		if err != nil {
			return false, fmt.Errorf("Failed scan table name, %w", err)
		}

		if name == table_name {
			has_table = true
			break
		}
	}

	return has_table, nil
}
