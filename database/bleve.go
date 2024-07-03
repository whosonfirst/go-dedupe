package database

// Remember: This assumes that Ollama is running in 'serve' mode on its default port.

// https://github.com/blevesearch/bleve/blob/c76f76d5176ed20783ec2fdcbc52642731fbe510/docs/vectors.md
// https://ollama.com/blog/embedding-models
// https://github.com/ollama/ollama/blob/main/docs/api.md

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/whosonfirst/go-dedupe/location"
)

type BleveDatabase struct {
	index bleve.Index
}

type BleveDocument struct {
	Id      string    `json:"id"`
	Text    string    `json:"text"`
	Geohash string    `json:"geohash"`
	Vec     []float32 `json:"vec"`
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

	q := u.Query()

	str_dims := q.Get("dimensions")

	dims, err := strconv.Atoi(str_dims)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse dimensions, %w", err)
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
		vectorFieldMapping := mapping.NewVectorFieldMapping()
		vectorFieldMapping.Dims = dims
		vectorFieldMapping.Similarity = "l2_norm" // euclidean distance

		bleveMapping := bleve.NewIndexMapping()
		bleveMapping.DefaultMapping.Dynamic = false
		bleveMapping.DefaultMapping.AddFieldMappingsAt("name", textFieldMapping)
		bleveMapping.DefaultMapping.AddFieldMappingsAt("vec", vectorFieldMapping)

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

func (db *BleveDatabase) Add(ctx context.Context, loc *location.Location) error {

	id := loc.ID
	text := fmt.Sprintf("%s, %s", loc.Name, loc.Address)

	doc := BleveDocument{
		Id:      id,
		Text:    text,
		Geohash: loc.Geohash(),
	}

	return db.index.Index(doc.Id, doc)
}

func (db *BleveDatabase) Query(ctx context.Context, loc *location.Location) ([]*QueryResult, error) {

	results := make([]*QueryResult, 0)
	return results, nil
}

func (db *BleveDatabase) Flush(ctx context.Context) error {
	return nil
}

func (db *BleveDatabase) Close(ctx context.Context) error {
	return nil
}
