package database

// https://opensearch.org/docs/latest/search-plugins/semantic-search/
// https://opensearch.org/docs/latest/field-types/supported-field-types/knn-vector/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v2/opensearchutil"
	"github.com/whosonfirst/go-whosonfirst-opensearch/client"
)

type opensearchDocument struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type OpensearchDatabase struct {
	Database
	client    *opensearch.Client
	index     string
	indexer   opensearchutil.BulkIndexer
	model_id  string
	waitGroup *sync.WaitGroup
}

func init() {
	ctx := context.Background()
	err := RegisterDatabase(ctx, "opensearch", NewOpensearchDatabase)

	if err != nil {
		panic(err)
	}
}

func NewOpensearchDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	dsn := q.Get("dsn")

	if dsn == "" {
		return nil, fmt.Errorf("Missing ?dsn= parameter")
	}

	// START OF put all of this in the go-whosonfirst-opensearch package

	os_client, err := client.NewClient(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create opensearch client, %w", err)
	}

	dsn_u, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse dsn (%s), %w", dsn, err)
	}

	os_index := dsn_u.Path
	os_index = strings.TrimLeft(os_index, "/")

	if os_index == "" {
		return nil, fmt.Errorf("dsn is missing ?index= parameter, '%s'", dsn)
	}

	wg := new(sync.WaitGroup)

	db := &OpensearchDatabase{
		client:    os_client,
		index:     os_index,
		waitGroup: wg,
	}

	bulk_index := true

	q_bulk_index := q.Get("bulk-index")

	if q_bulk_index != "" {

		v, err := strconv.ParseBool(q_bulk_index)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?bulk-index= parameter, %w", err)
		}

		bulk_index = v
	}

	if bulk_index {

		workers := 10

		q_workers := q.Get("workers")

		if q_workers != "" {

			w, err := strconv.Atoi(q_workers)

			if err != nil {
				return nil, fmt.Errorf("Failed to parse ?workers= parameter, %w", err)
			}

			workers = w
		}

		bi_cfg := opensearchutil.BulkIndexerConfig{
			Index:         os_index,
			Client:        os_client,
			NumWorkers:    workers,
			FlushInterval: 30 * time.Second,
			OnError: func(context.Context, error) {
				if err != nil {
					slog.Error("Bulk indexer reported an error", "error", err)
				}
			},
			// OnFlushStart func(context.Context) context.Context // Called when the flush starts.
			OnFlushEnd: func(context.Context) {
				slog.Debug("Bulk indexer flush end")
			},
		}

		bi, err := opensearchutil.NewBulkIndexer(bi_cfg)

		if err != nil {
			return nil, fmt.Errorf("Failed to create bulk indexer, %w", err)
		}

		db.indexer = bi
	}

	// END OF put all of this in the go-whosonfirst-opensearch package

	return db, nil
}

func (db *OpensearchDatabase) Add(ctx context.Context, id string, text string, metadata map[string]string) error {

	doc_id := id

	doc := opensearchDocument{
		ID:       id,
		Content:  text,
		Metadata: metadata,
	}

	enc_doc, err := json.Marshal(doc)

	if err != nil {
		return err
	}

	doc_r := bytes.NewReader(enc_doc)

	if db.indexer == nil {

		db.waitGroup.Add(1)
		defer db.waitGroup.Done()

		req := opensearchapi.IndexRequest{
			Index:      db.index,
			DocumentID: doc_id,
			Body:       doc_r,
			Refresh:    "true",
		}

		rsp, err := req.Do(ctx, db.client)

		if err != nil {
			return fmt.Errorf("Error getting response: %w", err)
		}

		defer rsp.Body.Close()

		if rsp.IsError() {
			return fmt.Errorf("Failed to index document, %v", rsp.Status())
		}

		return nil
	}

	// Do bulk index

	db.waitGroup.Add(1)

	bulk_item := opensearchutil.BulkIndexerItem{
		Action:     "index",
		DocumentID: doc_id,
		Body:       doc_r,

		OnSuccess: func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem) {
			slog.Debug("Index complete", "doc_id", doc_id)
			db.waitGroup.Done()
		},

		OnFailure: func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem, err error) {
			if err != nil {
				slog.Error("Failed to index record", "error", err)
			} else {
				slog.Error("Failed to index record", "type", res.Error.Type, "reason", res.Error.Reason)
			}

			db.waitGroup.Done()
		},
	}

	err = db.indexer.Add(ctx, bulk_item)

	if err != nil {
		return fmt.Errorf("Failed to add bulk item for %s, %w", doc_id, err)
	}

	return nil

}

func (db *OpensearchDatabase) Query(ctx context.Context, text string) ([]*QueryResult, error) {

	q := fmt.Sprintf(`{ "query": { "neural": { "content_embedding": { "query_text": "%s", "model_id": "%s", "k": 100 } } } }`, text, db.model_id)

	req := &opensearchapi.SearchRequest{
		Body: strings.NewReader(q),
		Index: []string{
			db.index,
		},
	}

	rsp, err := req.Do(ctx, db.client)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		slog.Error("Query failed", "status", rsp.StatusCode)
		return nil, fmt.Errorf("Invalid status")
	}

	body, err := io.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	slog.Info("Results", "body", string(body))

	results := make([]*QueryResult, 0)
	return results, nil
}

func (db *OpensearchDatabase) Flush(ctx context.Context) error {
	db.waitGroup.Wait()
	return nil
}

/*

"query": {
              "neural": {
                "passage_embedding": {
                  "query_text": "Hi world",
                  "model_id": "bQ1J8ooBpBj3wT4HVUsb",
                  "k": 100
                }
              }
            },
*/
