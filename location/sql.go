package location

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "log/slog"
	"net/url"
)

type SQLDatabase struct {
	conn   *sql.DB
	engine string
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

	// Something something something create table here if necessary...

	db := &SQLDatabase{
		engine: engine,
		conn:   conn,
	}

	if engine == "sqlite3" {

		err := db.configureSQLite(ctx)

		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (db *SQLDatabase) AddLocation(ctx context.Context, loc *Location) error {

	id := loc.ID
	geohash := loc.Geohash()

	enc_loc, err := json.Marshal(loc)

	if err != nil {
		return fmt.Errorf("Failed to marshal location, %w", err)
	}

	q := "REPLACE INTO locations (`id`, `geohash`, `body`) VALUES (?, ?, ?)"

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

func (db *SQLDatabase) GetWithGeohash(ctx context.Context, geohash string, cb GetWithGeohashCallback) error {

	q := "SELECT body FROM locations WHERE geohash = ?"

	rows, err := db.conn.QueryContext(ctx, q, geohash)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {

		var body []byte

		err := rows.Scan(&body)

		if err != nil {
			return err
		}

		var loc *Location

		err = json.Unmarshal(body, &loc)

		if err != nil {
			return err
		}

		err = cb(ctx, loc)

		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func (db *SQLDatabase) Close(ctx context.Context) error {
	return db.conn.Close()
}

func (db *SQLDatabase) configureSQLite(ctx context.Context) error {

	pragma := []string{
		"PRAGMA JOURNAL_MODE=OFF",
		"PRAGMA SYNCHRONOUS=OFF",
		"PRAGMA LOCKING_MODE=EXCLUSIVE",
		// https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html
		"PRAGMA PAGE_SIZE=4096",
		"PRAGMA CACHE_SIZE=1000000",
	}

	for _, p := range pragma {

		_, err := db.conn.ExecContext(ctx, p)

		if err != nil {
			return fmt.Errorf("Failed to set pragma '%s', %w", p, err)
		}
	}

	db.conn.SetMaxOpenConns(1)

	return nil
}
