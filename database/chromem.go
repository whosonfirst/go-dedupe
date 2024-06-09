package database

// Remember: This assumes that Ollama is running in 'serve' mode on its default port.

// https://github.com/philippgille/chromem-go
// https://ollama.com/blog/embedding-models
// https://github.com/ollama/ollama/blob/main/docs/api.md

import (
	"context"
	"fmt"
	"net/url"
	"runtime"

	"github.com/philippgille/chromem-go"
)

type ChromemDatabase struct {
	collection *chromem.Collection
	foo        int
}

func init() {
	ctx := context.Background()
	err := RegisterDatabase(ctx, "chromem", NewChromemDatabase)

	if err != nil {
		panic(err)
	}
}

func NewChromemDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	col_name := u.Host
	db_path := u.Path

	q := u.Query()
	model := q.Get("model")

	chromem_db, err := chromem.NewPersistentDB(db_path, false)

	if err != nil {
		return nil, fmt.Errorf("Failed to create database, %w", err)
	}

	collection, err := chromem_db.GetOrCreateCollection(col_name, nil, chromem.NewEmbeddingFuncOllama(model, ""))

	db := &ChromemDatabase{
		collection: collection,
		foo:        5,
	}

	return db, nil
}

func (db *ChromemDatabase) Add(ctx context.Context, id string, text string, metadata map[string]string) error {

	doc := chromem.Document{
		ID:      id,
		Content: text,
	}

	if metadata != nil {
		doc.Metadata = metadata
	}

	docs := []chromem.Document{
		doc,
	}

	return db.collection.AddDocuments(ctx, docs, runtime.NumCPU())
}

func (db *ChromemDatabase) Query(ctx context.Context, text string) ([]*QueryResult, error) {

	rsp, err := db.collection.Query(ctx, text, db.foo, nil, nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to query, %w", err)
	}

	results := make([]*QueryResult, len(rsp))

	for idx, r := range rsp {

		results[idx] = &QueryResult{
			ID:         r.ID,
			Metadata:   r.Metadata,
			Content:    r.Content,
			Embedding:  r.Embedding,
			Similarity: r.Similarity,
		}
	}

	return results, nil
}
