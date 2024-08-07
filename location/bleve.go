package location

import (
	"context"
	"encoding/json"
	"fmt"
	_ "log/slog"
	"net/url"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search"
)

type BleveDatabase struct {
	index bleve.Index
}

type bleveDocument struct {
	Id       string `json:"id"`
	Geohash  string `json:"geohash"`
	Location string `json:"location"`
}

func init() {
	ctx := context.Background()
	err := RegisterDatabase(ctx, "bleve", NewBleveDatabase)

	if err != nil {
		panic(err)
	}
}

func NewBleveDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	var idx bleve.Index
	db_path := u.Path

	_, err = os.Stat(db_path)

	if err == nil {

		v, err := bleve.Open(db_path)

		if err != nil {
			return nil, fmt.Errorf("Failed to open '%s', %w", db_path, err)
		}

		idx = v

	} else {

		textFieldMapping := mapping.NewTextFieldMapping()

		bleveMapping := bleve.NewIndexMapping()
		bleveMapping.DefaultMapping.AddFieldMappingsAt("geohash", textFieldMapping)

		v, err := bleve.New(db_path, bleveMapping)

		if err != nil {
			return nil, fmt.Errorf("Failed to create db at '%s', %w", db_path, err)
		}

		idx = v
	}

	db := &BleveDatabase{
		index: idx,
	}

	return db, nil
}

func (db *BleveDatabase) AddLocation(ctx context.Context, loc *Location) error {

	id := loc.ID
	geohash := loc.Geohash()

	enc_loc, err := json.Marshal(loc)

	if err != nil {
		return fmt.Errorf("Failed to marshal location, %w", err)
	}

	doc := bleveDocument{
		Id:       id,
		Geohash:  geohash,
		Location: string(enc_loc),
	}

	// slog.Info("Add", "id", loc.ID, "geohash", geohash)
	return db.index.Index(doc.Id, doc)
}

func (db *BleveDatabase) GetById(ctx context.Context, id string) (*Location, error) {

	q := bleve.NewDocIDQuery([]string{id})

	req := bleve.NewSearchRequest(q)
	req.Fields = []string{"location"}

	rsp, err := db.index.Search(req)

	if err != nil {
		return nil, err
	}

	if rsp.Total == 0 {
		return nil, nil
	}

	loc, err := db.locationFromDocumentMatch(rsp.Hits[0])

	if err != nil {
		return nil, err
	}

	return loc, nil
}

func (db *BleveDatabase) GetGeohashes(ctx context.Context, cb GetGeohashesCallback) error {
	return fmt.Errorf("Not implemeted")
}

func (db *BleveDatabase) GetWithGeohash(ctx context.Context, geohash string, cb GetWithGeohashCallback) error {

	q := bleve.NewMatchQuery(geohash)

	req := bleve.NewSearchRequest(q)
	req.Fields = []string{"location"}

	rsp, err := db.index.Search(req)

	if err != nil {
		return err
	}

	for _, m := range rsp.Hits {

		loc, err := db.locationFromDocumentMatch(m)

		if err != nil {
			return err
		}

		if loc.Geohash() != geohash {
			continue
		}

		err = cb(ctx, loc)

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *BleveDatabase) Close(ctx context.Context) error {
	return nil
}

func (db *BleveDatabase) locationFromDocumentMatch(m *search.DocumentMatch) (*Location, error) {

	fields := m.Fields

	v, exists := fields["location"]

	if !exists {
		return nil, fmt.Errorf("Match is missing location field")
	}

	enc_loc := v.(string)
	var loc *Location

	err := json.Unmarshal([]byte(enc_loc), &loc)

	if err != nil {
		return nil, err
	}

	return loc, nil
}
