package vector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/bwmarrin/snowflake"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whosonfirst/go-dedupe/embeddings"
	"github.com/whosonfirst/go-dedupe/location"
)

type SQLiteDatabase struct {
	vec_db     *sql.DB
	embedder   embeddings.Embedder
	dimensions int
}

var snowflake_node *snowflake.Node

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

	dimensions := 768

	if q.Has("dimensions") {

		v, err := strconv.Atoi(q.Get("dimensions"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?dimensions= paramter, %w", err)
		}

		dimensions = v
	}

	if snowflake_node == nil {

		n, err := snowflake.NewNode(1)

		if err != nil {
			return nil, fmt.Errorf("Failed to create snowflake node, %w", err)
		}

		snowflake_node = n
	}

	// See this? This important and without it none of the vec functions
	// will be registed
	sqlite_vec.Auto()

	vec_db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	has_meta_table, err := hasTable(ctx, vec_db, "vec_meta")

	if err != nil {
		return nil, fmt.Errorf("Failed to determine if metadata table exists, %w", err)
	}

	if !has_meta_table {

		q := "CREATE TABLE vec_meta (id TEXT PRIMARY KEY, snowflake_id INTEGER); CREATE INDEX `vec_meta_by_snowflake_id` ON vec_meta (`snowflake_id`)"
		slog.Debug(q)

		_, err := vec_db.ExecContext(ctx, q)

		if err != nil {
			return nil, fmt.Errorf("Failed to create vec_meta table, %w", err)
		}

	}

	has_vec_table, err := hasTable(ctx, vec_db, "vec_items")

	if err != nil {
		return nil, fmt.Errorf("Failed to determine if vec_items table exists, %w", err)
	}

	if !has_vec_table {

		q := fmt.Sprintf("CREATE VIRTUAL TABLE vec_items USING vec0(embedding float[%d])", dimensions)
		slog.Debug(q)

		_, err := vec_db.ExecContext(ctx, q)

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
		vec_db:     vec_db,
		embedder:   embdr,
		dimensions: dimensions,
	}

	return db, nil
}

func (db *SQLiteDatabase) Add(ctx context.Context, loc *location.Location) error {

	id := loc.ID

	snowflake_id, err := db.getSnowflakeId(ctx, loc)

	if err != nil {
		return err
	}

	v, err := db.embeddings(ctx, loc)

	if err != nil {
		return fmt.Errorf("Failed to serialize floats for ID %s, %w", id, err)
	}

	_, err = db.vec_db.ExecContext(ctx, "INSERT INTO vec_items(rowid, embedding) VALUES (?, ?)", snowflake_id, v)

	if err != nil {
		return fmt.Errorf("Failed to insert row for ID %s (%d), %w", id, snowflake_id, err)
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

		var snowflake_id int64
		var distance float64

		err = rows.Scan(&snowflake_id, &distance)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan row, %w", err)
		}

		id, err := db.getLocationId(ctx, snowflake_id)

		if err != nil {
			return nil, err
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

func (db *SQLiteDatabase) getSnowflakeId(ctx context.Context, loc *location.Location) (int64, error) {

	q := "SELECT snowflake_id FROM vec_meta WHERE id = ?"
	row := db.vec_db.QueryRowContext(ctx, q, loc.ID)

	var snowflake_id int64

	err := row.Scan(&snowflake_id)

	switch {
	case err == sql.ErrNoRows:

		new_id := snowflake_node.Generate()
		snowflake_id = new_id.Int64()

		q := "INSERT INTO vec_meta (id, snowflake_id) VALUES(?, ?)"

		_, err := db.vec_db.ExecContext(ctx, q, loc.ID, snowflake_id)

		if err != nil {
			return 0, fmt.Errorf("Failed to create entry for snowflake ID, %w", err)
		}

		return snowflake_id, nil

	case err != nil:
		return 0, fmt.Errorf("Failed to retrieve snowflake ID, %w", err)
	default:
		return snowflake_id, nil
	}
}

func (db *SQLiteDatabase) getLocationId(ctx context.Context, snowflake_id int64) (string, error) {

	q := "SELECT id FROM vec_meta WHERE snowflake_id = ?"
	row := db.vec_db.QueryRowContext(ctx, q, snowflake_id)

	var id string

	err := row.Scan(&id)

	if err != nil {
		return "", fmt.Errorf("Failed to retrieve location ID, %w", err)
	}

	return id, nil
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
