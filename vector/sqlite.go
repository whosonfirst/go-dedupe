package vector

// https://alexgarcia.xyz/blog/2024/sqlite-vec-stable-release/index.html
// https://alexgarcia.xyz/sqlite-vec/go.html

// https://github.com/asg017/sqlite-lembed

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"
	
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/bwmarrin/snowflake"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whosonfirst/go-dedupe/embeddings"
	"github.com/whosonfirst/go-dedupe/location"
)

type SQLiteDatabase struct {
	// The underlying SQLite database used to store and query embeddings.
	vec_db       *sql.DB
	// The whosonfirst/go-dedupe/embeddings instance to use for deriving embeddings.
	embedder     embeddings.Embedder
	// The number of dimensions for embeddings
	dimensions   int
	// The maximum distance between query input and embeddings when matching
	max_distance float32
	// The maximum number of results for queries
	max_results  int
	// The compression type to use for embeddings. Valid options are: quantize, matroyshka, none (default)
	compression  string
	// If true that existing records are re-indexed. If not, they are skipped and left as-is.
	refresh bool
}

var snowflake_node *snowflake.Node

const matroyshka_dimensions int = 512

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
	max_distance := float32(5.0)
	max_results := 10
	compression := "none"
	refresh := false
	
	if q.Has("dimensions") {

		v, err := strconv.Atoi(q.Get("dimensions"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?dimensions= parameter, %w", err)
		}

		dimensions = v
	}

	if q.Has("max-distance") {

		v, err := strconv.ParseFloat(q.Get("max-distance"), 64)

		if err != nil {
			return nil, fmt.Errorf("Invalid ?max-distance= parameter, %w", err)
		}

		max_distance = float32(v)
	}

	if q.Has("max-results") {

		v, err := strconv.Atoi(q.Get("max-results"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?max-results= parameter, %w", err)
		}

		max_results = v
	}

	if q.Has("compression") {
		compression = q.Get("compression")
	}

	if q.Has("refresh"){

		v, err := strconv.ParseBool(q.Get("refresh"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?refresh= parameter, %w", err)
		}

		refresh = v
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

		q := "CREATE TABLE vec_meta (id TEXT PRIMARY KEY, snowflake_id INTEGER, content TEXT); CREATE INDEX `vec_meta_by_snowflake_id` ON vec_meta (`snowflake_id`);"
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

		var q string

		switch compression {
		case "quantize":
			q = fmt.Sprintf("CREATE VIRTUAL TABLE vec_items USING vec0(embedding bit[%d])", dimensions)
		case "matroyshka":
			q = fmt.Sprintf("CREATE VIRTUAL TABLE vec_items USING vec0(embedding float[%d])", matroyshka_dimensions)
		case "none":
			q = fmt.Sprintf("CREATE VIRTUAL TABLE vec_items USING vec0(embedding float[%d])", dimensions)
		default:
			return nil, fmt.Errorf("Invalid or unsupported compression")
		}

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
		vec_db:       vec_db,
		embedder:     embdr,
		dimensions:   dimensions,
		max_distance: max_distance,
		max_results:  max_results,
		compression:  compression,
		refresh: refresh,
	}

	return db, nil
}

func (db *SQLiteDatabase) Add(ctx context.Context, loc *location.Location) error {

	id := loc.ID

	snowflake_id, err := db.getSnowflakeId(ctx, loc)

	if err != nil {
		return err
	}

	// START OF UPSERT not implemented for virtual table "vec_items"

	action := "insert"

	q := "SELECT rowid FROM vec_items WHERE rowid = ?"
	row := db.vec_db.QueryRowContext(ctx, q, snowflake_id)

	var rowid int64
	err = row.Scan(&rowid)

	switch {
	case err == sql.ErrNoRows:
		// pass
	case err != nil:
		return fmt.Errorf("Failed to determine if rowid (%d) exists, %w", snowflake_id, err)
	default:

		action = "skip"
		
		if db.refresh {
			action = "update"
		}
	}

	// END OF UPSERT not implemented for virtual table "vec_items"

	slog.Debug("Add embeddings", "id", loc.ID, "snowflake id", snowflake_id, "rowid", rowid, "action", action)

	switch action {
	case "skip":
		return nil
	case "insert", "update":

		v, err := db.embeddings(ctx, loc)

		if err != nil {
			return fmt.Errorf("Failed to serialize floats for ID %s, %w", id, err)
		}

		switch action {
		case "update":

			var q string

			switch db.compression {
			case "quantize":
				q = "UPDATE vec_items SET vec_quantize_binary(embedding) = ? WHERE rowid = ?"
			case "matroyshka":
				q = fmt.Sprintf("UPDATE vec_items SET vec_normalize(vec_slice(eembedding, 0, %d)) = ? WHERE rowid = ?", matroyshka_dimensions)
			case "none":
				q = "UPDATE vec_items SET embedding = ? WHERE rowid = ?"
			default:
				return fmt.Errorf("Invalid or unsupported compression")
			}

			slog.Debug(q)

			_, err = db.vec_db.ExecContext(ctx, q, v, snowflake_id)

			if err != nil {
				return fmt.Errorf("Failed to update row for ID %s (%d), %w", id, snowflake_id, err)
			}

		default:

			var q string

			switch db.compression {
			case "quantize":
				q = "INSERT INTO vec_items(rowid, embedding) VALUES (?, vec_quantize_binary(?))"
			case "matroyshka":
				q = fmt.Sprintf("INSERT INTO vec_items(rowid, embedding) VALUES (?, vec_normalize(vec_slice(?, 0, %d)))", matroyshka_dimensions)
			case "none":
				q = "INSERT INTO vec_items(rowid, embedding) VALUES (?, ?)"
			default:
				return fmt.Errorf("Invalid or unsupported compression")
			}

			slog.Debug(q)

			_, err := db.vec_db.ExecContext(ctx, q, snowflake_id, v)

			if err != nil {
				return fmt.Errorf("Failed to insert row for ID %s (%d), %w", id, snowflake_id, err)
			}
		}

	default:
		return fmt.Errorf("Invalid or unsupported action")
	}

	return nil
}

func (db *SQLiteDatabase) Query(ctx context.Context, loc *location.Location) ([]*QueryResult, error) {

	results := make([]*QueryResult, 0)

	query, err := db.embeddings(ctx, loc)

	if err != nil {
		return nil, fmt.Errorf("Failed to serialize query, %w", err)
	}

	var q string

	switch db.compression {
	case "quantize":
		q = "SELECT rowid, distance FROM vec_items WHERE embedding MATCH vec_quantize_binary(?)"
	case "matroyshka":
		q = fmt.Sprintf("SELECT rowid, distance FROM vec_items WHERE embedding MATCH vec_normalize(vec_slice(?, 0, %d))", matroyshka_dimensions)
	case "none":
		q = "SELECT rowid, distance FROM vec_items WHERE embedding MATCH ?"
	default:
		return nil, fmt.Errorf("Invalid or unsupported compression")
	}

	q = fmt.Sprintf("%s AND distance <= ? ORDER BY distance LIMIT ?", q)

	slog.Debug("Query", "statement", q, "location", loc, "distance", db.max_distance, "limit", db.max_results, "compression", db.compression)

	t1 := time.Now()
	
	rows, err := db.vec_db.QueryContext(ctx, q, query, db.max_distance, db.max_results)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute query, %w", err)
	}

	slog.Debug("Query context", "time", time.Since(t1))
	
	for rows.Next() {

		var snowflake_id int64
		var distance float64

		err = rows.Scan(&snowflake_id, &distance)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan row, %w", err)
		}

		slog.Debug("Query scan", "id", snowflake_id, "time", time.Since(t1))
		
		id, content, err := db.getLocationData(ctx, snowflake_id)

		if err != nil {
			return nil, err
		}

		r := &QueryResult{
			ID:         id,
			Content:    content,
			Similarity: float32(distance),
		}

		slog.Debug("Result", "rowid", snowflake_id, "location id", id, "content", content, "distance", distance)

		results = append(results, r)
	}

	slog.Debug("Query rows", "time", time.Since(t1))
	
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

		q := "INSERT INTO vec_meta (id, snowflake_id, content) VALUES(?, ?, ?)"

		_, err := db.vec_db.ExecContext(ctx, q, loc.ID, snowflake_id, loc.String())

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

func (db *SQLiteDatabase) getLocationData(ctx context.Context, snowflake_id int64) (string, string, error) {

	q := "SELECT id, content FROM vec_meta WHERE snowflake_id = ?"
	row := db.vec_db.QueryRowContext(ctx, q, snowflake_id)

	var id string
	var content string

	err := row.Scan(&id, &content)

	if err != nil {
		return "", "", fmt.Errorf("Failed to retrieve location ID, %w", err)
	}

	return id, content, nil
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
