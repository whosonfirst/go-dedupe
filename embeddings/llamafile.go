//go:build llamafile

package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type LlamafileEmbeddingRequest struct {
	Content string `json:"content,omitempty"`
}

type LlamafileEmbeddingResponse struct {
	Embeddings []float64 `json:"embedding,omitempty"`
}

// LlamafileEmbedder implements the `Embedder` interface using an Llamafile API endpoint to derive embeddings.
type LlamafileEmbedder struct {
	Embedder
	client *http.Client
	model  string
}

func init() {
	ctx := context.Background()
	err := RegisterEmbedder(ctx, "llamafile", NewLlamafileEmbedder)

	if err != nil {
		panic(err)
	}
}

func NewLlamafileEmbedder(ctx context.Context, uri string) (Embedder, error) {

	_, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	// q := u.Query()

	cl := &http.Client{}

	e := &LlamafileEmbedder{
		client: cl,
	}

	return e, nil
}

func (e *LlamafileEmbedder) Embeddings(ctx context.Context, content string) ([]float64, error) {

	// curl --request POST --url http://localhost:8080/embedding --header "Content-Type: application/json" --data '{"content": "Hello world" }'

	endpoint := "http://localhost:8080/embedding"

	msg := LlamafileEmbeddingRequest{
		Content: content,
	}

	enc_msg, err := json.Marshal(msg)

	if err != nil {
		return nil, fmt.Errorf("Failed to encode message, %w", err)
	}

	br := bytes.NewReader(enc_msg)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, br)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new request, %w", err)
	}

	req.Header.Set("Content-type", "application/json")

	rsp, err := e.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute request, %w", err)
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Embeddings request failed %d: %s", rsp.StatusCode, rsp.Status)
	}

	var llamafile_rsp *LlamafileEmbeddingResponse

	dec := json.NewDecoder(rsp.Body)
	err = dec.Decode(&llamafile_rsp)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal embeddings, %w", err)
	}

	return llamafile_rsp.Embeddings, nil
}

// TBD...

func (e *LlamafileEmbedder) Embeddings32(ctx context.Context, content string) ([]float32, error) {

	e64, err := e.Embeddings(ctx, content)

	if err != nil {
		return nil, err
	}

	e32 := make([]float32, len(e64))

	for idx, v := range e64 {
		e32[idx] = float32(v)
	}

	return e32, nil
}
