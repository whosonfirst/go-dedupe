package vector

// https://opensearch.org/docs/latest/search-plugins/semantic-search/
// https://opensearch.org/docs/latest/field-types/supported-field-types/knn-vector/
// https://opensearch.org/docs/latest/search-plugins/knn/settings/

// https://opensearch.org/blog/improving-document-retrieval-with-sparse-semantic-encoders/
// https://opensearch.org/blog/A-deep-dive-into-faster-semantic-sparse-retrieval-in-OS-2.12/

// https://opensearch.org/docs/latest/search-plugins/knn/filter-search-knn/
// https://opensearch.org/docs/latest/field-types/supported-field-types/knn-vector/

// https://junming-chen.medium.com/using-elasticsearch-as-a-vector-database-dive-into-dense-vector-and-script-score-198e2eb807d6
// https://www.elastic.co/search-labs/blog/how-to-deploy-nlp-text-embeddings-and-vector-search

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v2/opensearchutil"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-whosonfirst-opensearch/client"
)

//go:embed opensearch_*.tpl
var opensearch_fs embed.FS

type opensearchDocument struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type opensearchQueryVars struct {
	Location *location.Location
	ModelId  string
}

type OpensearchDatabase struct {
	Database
	client          *opensearch.Client
	index           string
	indexer         opensearchutil.BulkIndexer
	model_id        string
	waitGroup       *sync.WaitGroup
	query_templates *template.Template
	query_label     string
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

	// START OF put all of this in the go-whosonfirst-opensearch package

	dsn := q.Get("dsn")

	if dsn == "" {
		return nil, fmt.Errorf("Missing ?dsn= parameter")
	}

	model := q.Get("model")

	if model == "" {
		return nil, fmt.Errorf("Missing ?model= parameter")
	}

	os_client, err := client.NewClient(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create opensearch client, %w", err)
	}

	dsn_u, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse dsn (%s), %w", dsn, err)
	}

	// This assumes that the relevant index has already been created and configured
	// to use an ML/embedding model. But what if did all of that on-the-fly at runtime
	// here (and then tore down the index in the Close method below).

	os_index := dsn_u.Path
	os_index = strings.TrimLeft(os_index, "/")

	if os_index == "" {
		return nil, fmt.Errorf("dsn is missing ?index= parameter, '%s'", dsn)
	}

	t := template.New("opensearch").Funcs(template.FuncMap{
		"HasMetadata": func(m map[string]string) bool {
			has_metadata := false
			for range m {
				has_metadata = true
				break
			}
			return has_metadata
		},
	})

	t, err = t.ParseFS(opensearch_fs, "*.tpl")

	if err != nil {
		return nil, fmt.Errorf("Failed to parse query templates, %w", err)
	}

	// Read from query param...
	query_label := "opensearch_query_neural_text"
	// query_label := "opensearch_query_simple"

	wg := new(sync.WaitGroup)

	db := &OpensearchDatabase{
		client:          os_client,
		index:           os_index,
		model_id:        model,
		query_templates: t,
		query_label:     query_label,
		waitGroup:       wg,
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

func (db *OpensearchDatabase) Add(ctx context.Context, loc *location.Location) error {

	doc_id := loc.ID

	logger := slog.Default()
	logger = logger.With("index", db.index)
	logger = logger.With("id", doc_id)

	doc := opensearchDocument{
		ID:       loc.ID,
		Name:     loc.Name,
		Address:  loc.Address,
		Metadata: loc.Metadata(),
	}

	// START OF put all of this in the go-whosonfirst-opensearch package

	enc_doc, err := json.Marshal(doc)

	if err != nil {
		logger.Error("Failed to marshal record", "error", err)
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
			logger.Error("Failed to execute indexing request", "status", rsp.Status)
			return fmt.Errorf("Error getting response: %w", err)
		}

		defer rsp.Body.Close()

		if rsp.IsError() {
			body, _ := io.ReadAll(rsp.Body)
			return fmt.Errorf("Failed to index document, %v, %s", rsp.Status(), string(body))
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

	// END OF put all of this in the go-whosonfirst-opensearch package

	return nil

}

func (db *OpensearchDatabase) Query(ctx context.Context, loc *location.Location) ([]*QueryResult, error) {

	t := db.query_templates.Lookup(db.query_label)

	if t == nil {
		return nil, fmt.Errorf("Missing opensearch_query template")
	}

	vars := opensearchQueryVars{
		Location: loc,
		ModelId:  db.model_id,
	}

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	err := t.Execute(wr, vars)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive query, %w", err)
	}

	wr.Flush()

	// fmt.Println(buf.String())

	body_r := bytes.NewReader(buf.Bytes())

	req := &opensearchapi.SearchRequest{
		Body: body_r,
		Index: []string{
			db.index,
		},
	}

	rsp, err := req.Do(ctx, db.client)

	if err != nil {
		slog.Error("Query request failed", "error", err)
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		body, _ := io.ReadAll(rsp.Body)
		slog.Error("Query execution failed", "status", rsp.StatusCode, "query", buf.String(), "response", string(body))
		return nil, fmt.Errorf("Invalid status")
	}

	body, err := io.ReadAll(rsp.Body)

	if err != nil {
		slog.Error("Failed to read response body", "error", err)
		return nil, err
	}

	// slog.Info("DEBUG", "body", string(body))

	hits_r := gjson.GetBytes(body, "hits.hits")
	count_hits := len(hits_r.Array())

	results := make([]*QueryResult, count_hits)

	for idx, r := range hits_r.Array() {

		score_rsp := r.Get("_score")
		score := score_rsp.Float()

		src := r.Get("_source")

		id := src.Get("id").String()
		name := src.Get("name").String()
		address := src.Get("address").String()

		content := fmt.Sprintf("%s %s", name, address)

		qr := &QueryResult{
			ID:         id,
			Content:    content,
			Similarity: float32(score),
		}

		results[idx] = qr
	}

	return results, nil
}

func (db *OpensearchDatabase) MeetsThreshold(ctx context.Context, qr *QueryResult, threshold float64) (bool, error) {

	if float64(qr.Similarity) > threshold {
		return false, nil
	}

	return true, nil
}

func (db *OpensearchDatabase) Flush(ctx context.Context) error {
	db.waitGroup.Wait()
	return nil
}

func (db *OpensearchDatabase) Close(ctx context.Context) error {

	// See notes in NewOpensearchDatabase about creating indices on the
	// fly and tearing them down here.

	return nil
}
