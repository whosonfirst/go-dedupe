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
	"github.com/tidwall/gjson"
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

	model := q.Get("model")

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
		model_id:  model,
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

	logger := slog.Default()
	logger = logger.With("index", db.index)
	logger = logger.With("id", id)

	doc_id := id

	doc := opensearchDocument{
		ID:       id,
		Content:  text,
		Metadata: metadata,
	}

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
			logger.Error("Failed to index record", "status", rsp.Status)
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

	//

	hits_r := gjson.GetBytes(body, "hits.hits")
	count_hits := len(hits_r.Array())

	slog.Info("COUNT", "hits", count_hits)

	for idx, r := range hits_r.Array() {

		score_rsp := r.Get("_score")
		score := score_rsp.Float()

		src := r.Get("_source")

		id := src.Get("id")
		content := src.Get("content")

		slog.Info("R", "index", idx, "score", score, "id", id, "content", content)
	}

	// slog.Info("Results", "body", string(body))

	results := make([]*QueryResult, 0)
	return results, nil
}

func (db *OpensearchDatabase) Flush(ctx context.Context) error {
	db.waitGroup.Wait()
	return nil
}

/*

[2024-06-10T23:02:35,762][WARN ][o.o.m.j.JvmGcMonitorService] [dd22e97b6dc8] [gc][2257] overhead, spent [1s] collecting in the last [1s]
java.lang.OutOfMemoryError: Java heap space
Dumping heap to data/java_pid43.hprof ...
Heap dump file created [1339409158 bytes in 5.371 secs]
[2024-06-10T23:02:41,864][WARN ][o.o.m.j.JvmGcMonitorService] [dd22e97b6dc8] [gc][2258] overhead, spent [3.7s] collecting in the last [6.1s]
[2024-06-10T23:02:41,868][ERROR][o.o.b.OpenSearchUncaughtExceptionHandler] [dd22e97b6dc8] fatal error in thread [opensearch[dd22e97b6dc8][write][T#2]], exiting
java.lang.OutOfMemoryError: Java heap space
[2024-06-10T23:02:41,868][ERROR][o.o.b.OpenSearchUncaughtExceptionHandler] [dd22e97b6dc8] fatal error in thread [opensearch[dd22e97b6dc8][generic][T#1]], exiting
java.lang.OutOfMemoryError: Java heap space
fatal error in thread [opensearch[dd22e97b6dc8][generic][T#1]], exiting
java.lang.OutOfMemoryError: Java heap space
[2024-06-10T23:02:41,867][ERROR][o.o.b.OpenSearchUncaughtExceptionHandler] [dd22e97b6dc8] fatal error in thread [opensearch[dd22e97b6dc8][write][T#9]], exiting
java.lang.OutOfMemoryError: Java heap space
fatal error in thread [opensearch[dd22e97b6dc8][write][T#9]], exiting
java.lang.OutOfMemoryError: Java heap space
fatal error in thread [opensearch[dd22e97b6dc8][write][T#2]], exiting
java.lang.OutOfMemoryError: Java heap space
[2024-06-10T23:02:41,866][ERROR][o.o.b.OpenSearchUncaughtExceptionHandler] [dd22e97b6dc8] fatal error in thread [opensearch[dd22e97b6dc8][write][T#7]], exiting
java.lang.OutOfMemoryError: Java heap space
	at java.base/java.util.concurrent.ThreadPoolExecutor.addWorker(ThreadPoolExecutor.java:928) ~[?:?]
	at java.base/java.util.concurrent.ThreadPoolExecutor.processWorkerExit(ThreadPoolExecutor.java:1021) ~[?:?]
	at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1158) ~[?:?]
	at java.base/java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:642) ~[?:?]
	at java.base/java.lang.Thread.runWith(Thread.java:1596) ~[?:?]
	at java.base/java.lang.Thread.run(Thread.java:1583) [?:?]
fatal error in thread [opensearch[dd22e97b6dc8][write][T#7]], exiting
[2024-06-10T23:02:41,868][ERROR][o.o.i.e.Engine           ] [dd22e97b6dc8] [dedupe][0] already closed by tragic event on the index writer
java.lang.OutOfMemoryError: Java heap space

*/
