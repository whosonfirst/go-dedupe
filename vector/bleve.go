//go:build vectors
// +build vectors

package vector

// Remember: This assumes that Ollama is running in 'serve' mode on its default port.

// https://github.com/blevesearch/bleve/blob/c76f76d5176ed20783ec2fdcbc52642731fbe510/docs/vectors.md
// https://ollama.com/blog/embedding-models
// https://github.com/ollama/ollama/blob/main/docs/api.md

/*

> go run --tags=vectors cmd/index-overture-places/main.go -database-uri 'bleve:///usr/local/data/venues-b.db?dimensions=768&embedder-uri=chromemollama://?model=mxbai-embed-large' /usr/local/data/overture/places-geojson/*.bz2

2024/07/03 15:29:44 ERROR Failed to add record path=usr/local/data/overture/places-geojson/venues-0.95.geojsonl.bz2 "line number"=13624 id="ovtr:id=08fa8b1a31a6150c0357ed79da9e1fc2" location="Vou de Marisa, Avenida Rui Barbosa, 597 Macaé RJ BR" error="Failed to derive embeddings, error response from the embedding API: 400 Bad Request"

*/

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/whosonfirst/go-dedupe/embeddings"
	"github.com/whosonfirst/go-dedupe/location"
)

type BleveDatabase struct {
	index    bleve.Index
	embedder embeddings.Embedder
	tmp_file string
}

type BleveDocument struct {
	Id         string    `json:"id"`
	Text       string    `json:"text"`
	Geohash    string    `json:"geohash"`
	Embeddings []float32 `json:"embeddings"`
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

	embedder_uri := q.Get("embedder-uri")

	embdr, err := embeddings.NewEmbedder(ctx, embedder_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new embedder, %w", err)
	}

	var idx bleve.Index
	db_path := u.Path

	db_path, tmp_file, err := EntempifyURI(db_path)

	if err != nil {
		return nil, fmt.Errorf("Failed to entempify URI, %w", err)
	}

	textFieldMapping := mapping.NewTextFieldMapping()
	vectorFieldMapping := mapping.NewVectorFieldMapping()
	vectorFieldMapping.Dims = dims
	vectorFieldMapping.Similarity = "l2_norm" // euclidean distance

	bleveMapping := bleve.NewIndexMapping()
	bleveMapping.DefaultMapping.Dynamic = false
	bleveMapping.DefaultMapping.AddFieldMappingsAt("name", textFieldMapping)
	bleveMapping.DefaultMapping.AddFieldMappingsAt("vec", vectorFieldMapping)

	idx, err := bleve.New(db_path, bleveMapping)

	if err != nil {
		return nil, fmt.Errorf("Failed to create db at '%s', %w", db_path, err)
	}

	db := &BleveDatabase{
		index:    idx,
		embedder: embdr,
		tmp_file: tmp_file,
	}

	return db, nil
}

func (db *BleveDatabase) Add(ctx context.Context, loc *location.Location) error {

	id := loc.ID
	text := fmt.Sprintf("%s, %s", loc.Name, loc.Address)

	embeddings, err := db.embedder.Embeddings32(ctx, text)

	if err != nil {
		return fmt.Errorf("Failed to derive embeddings for '%s', %w", text, err)
	}

	doc := BleveDocument{
		Id:         id,
		Text:       text,
		Geohash:    loc.Geohash(),
		Embeddings: embeddings,
	}

	return db.index.Index(doc.Id, doc)
}

func (db *BleveDatabase) Query(ctx context.Context, loc *location.Location) ([]*QueryResult, error) {

	text := fmt.Sprintf("%s, %s", loc.Name, loc.Address)

	embeddings, err := db.embedder.Embeddings32(ctx, text)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive embeddings, %w", err)
	}

	k := int64(5)
	boost := 0.0

	req := bleve.NewSearchRequest(bleve.NewMatchNoneQuery())
	req.AddKNN("embeddings", embeddings, k, boost)

	rsp, err := db.index.Search(req)

	if err != nil {
		return nil, err
	}

	slog.Info("Q", "hits", rsp.Hits)

	results := make([]*QueryResult, 0)
	return results, nil
}

func (db *BleveDatabase) Flush(ctx context.Context) error {
	return nil
}

func (db *BleveDatabase) Close(ctx context.Context) error {

	err := db.idx.Close()

	if err != nil {
		return err
	}

	if tmp_file != "" {
		err := os.Remove(tmp_file)

		if err != nil {
			return err
		}
	}

	return nil
}
