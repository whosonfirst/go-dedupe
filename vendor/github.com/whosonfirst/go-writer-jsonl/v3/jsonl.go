package jsonl

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/whosonfirst/go-writer/v3"
	"io"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
)

func init() {

	ctx := context.Background()

	err := writer.RegisterWriter(ctx, "jsonl", NewJSONLWriter)

	if err != nil {
		panic(err)
	}
}

type JSONLWriter struct {
	writer.Writer
	writer writer.Writer
	mu     *sync.RWMutex
	count  int64
}

func NewJSONLWriter(ctx context.Context, uri string) (writer.Writer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	wr_uri := q.Get("writer")

	if wr_uri == "" {
		return nil, fmt.Errorf("Missing ?writer= parameter")
	}

	wr, err := writer.NewWriter(ctx, wr_uri)

	if err != nil {
		return nil, err
	}

	mu := new(sync.RWMutex)

	jsonl_wr := &JSONLWriter{
		writer: wr,
		mu:     mu,
		count:  int64(0),
	}

	return jsonl_wr, nil
}

func (jsonl_wr *JSONLWriter) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	jsonl_wr.mu.Lock()

	defer func() {
		jsonl_wr.mu.Unlock()
		atomic.AddInt64(&jsonl_wr.count, 1)
	}()

	var doc interface{}

	dec := json.NewDecoder(fh)
	err := dec.Decode(&doc)

	if err != nil {
		return 0, fmt.Errorf("Failed to decode %s, %w", key, err)
	}

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	enc := json.NewEncoder(wr)
	err = enc.Encode(doc)

	if err != nil {
		return 0, fmt.Errorf("Failed to encode %s, %w", key, err)
	}

	wr.Flush()

	br := bytes.NewReader(buf.Bytes())
	return jsonl_wr.writer.Write(ctx, key, br)
}

func (jsonl_wr *JSONLWriter) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

func (jsonl_wr *JSONLWriter) Flush(ctx context.Context) error {
	return nil
}

func (jsonl_wr *JSONLWriter) Close(ctx context.Context) error {
	return nil
}

func (jsonl_wr *JSONLWriter) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
