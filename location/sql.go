package location

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/whosonfirst/go-dedupe/database"
)

type SQLDatabase struct {
	conn   *sql.DB
	engine string
	dsn    string
}

func init() {
	ctx := context.Background()
	err := RegisterDatabase(ctx, "sql", NewSQLDatabase)

	if err != nil {
		panic(err)
	}
}

func NewSQLDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	engine := u.Host

	q := u.Query()
	dsn := q.Get("dsn")

	conn, err := sql.Open(engine, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	db := &SQLDatabase{
		engine: engine,
		conn:   conn,
		dsn:    dsn,
	}

	opts := database.DefaultConfigureSQLDatabaseOptions()
	opts.CreateTablesIfNecessary = true

	opts.Tables = []*database.SQLTable{
		&database.SQLTable{
			Name:   "locations",
			Schema: "CREATE TABLE locations (id TEXT PRIMARY KEY, geohash TEXT, body TEXT); CREATE INDEX locations_by_geohash ON locations (geohash);",
		},
	}

	err = database.ConfigureSQLDatabase(ctx, db.conn, opts)

	if err != nil {
		return nil, err
	}

	switch engine {
	case "sqlite3":

		pragma := database.DefaultSQLitePragma()
		err := database.ConfigureSQLitePragma(ctx, db.conn, pragma)

		if err != nil {
			return nil, err
		}
	default:
		//
	}

	if q.Has("max-conns") {

		v, err := strconv.Atoi(q.Get("max-conns"))

		if err != nil {
			return nil, err
		}

		db.conn.SetMaxOpenConns(v)
	}

	return db, nil
}

func (db *SQLDatabase) String() string {
	return db.dsn
}

func (db *SQLDatabase) AddLocation(ctx context.Context, loc *Location) error {

	id := loc.ID
	geohash := loc.Geohash()

	enc_loc, err := json.Marshal(loc)

	if err != nil {
		return fmt.Errorf("Failed to marshal location, %w", err)
	}

	q := "INSERT OR REPLACE INTO locations (id, geohash, body) VALUES (?, ?, ?)"

	_, err = db.conn.ExecContext(ctx, q, id, geohash, string(enc_loc))

	if err != nil {
		return fmt.Errorf("Failed to add location, %w", err)
	}

	return nil
}

func (db *SQLDatabase) GetById(ctx context.Context, id string) (*Location, error) {

	q := "SELECT body FROM locations WHERE id = ?"

	row := db.conn.QueryRowContext(ctx, q, id)

	var body []byte

	err := row.Scan(&body)

	if err != nil {
		return nil, err
	}

	var loc *Location

	err = json.Unmarshal(body, &loc)

	if err != nil {
		return nil, err
	}

	return loc, nil
}

func (db *SQLDatabase) GetGeohashes(ctx context.Context) iter.Seq2[string, error] {

	return func(yield func(string, error) bool) {

		// To do: Make ASC / DESC a config option

		q := "SELECT geohash, COUNT(id) AS count FROM locations GROUP BY geohash ORDER BY count DESC"
		slog.Debug("Get geohashes", "query", q)

		rows, err := db.conn.QueryContext(ctx, q)

		if err != nil {
			yield("", err)
			return
		}

		defer rows.Close()

		for rows.Next() {

			var geohash string
			var count int

			err := rows.Scan(&geohash, &count)

			if err != nil {
				yield("", err)
				return
			}

			slog.Debug("Handle geohash", "geohash", geohash)

			if !yield(geohash, nil) {
				break
			}
		}

		return
	}

}

func (db *SQLDatabase) GetWithGeohash(ctx context.Context, geohash string) iter.Seq2[*Location, error] {

	return func(yield func(*Location, error) bool) {

		q := "SELECT body FROM locations WHERE geohash = ?"
		slog.Debug("Get with geohash", "query", q, "geohash", geohash, "database", db)

		rows, err := db.conn.QueryContext(ctx, q, geohash)

		if err != nil {
			slog.Error("Failed to query", "error", err)
			yield(nil, err)
			return
		}

		defer rows.Close()

		for rows.Next() {

			var body []byte

			err := rows.Scan(&body)

			if err != nil {
				slog.Error("Failed to scan location body", "error", err)
				yield(nil, err)
				return
			}

			var loc *Location

			err = json.Unmarshal(body, &loc)

			if err != nil {
				slog.Error("Failed to unmarshal location", "error", err)
				yield(nil, err)
				return
			}

			slog.Debug("Process location for geohash", "geohash", geohash, "location", loc.String())

			if !yield(loc, err) {
				break
			}
		}

		err = rows.Err()

		if err != nil {
			yield(nil, err)
		}

		return
	}
}

func (db *SQLDatabase) Close(ctx context.Context) error {
	return db.conn.Close()
}
