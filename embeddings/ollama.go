//go:build ollama

package embeddings

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

// OllamaEmbedder implements the `Embedder` interface using an Ollama API endpoint to derive embeddings.
type OllamaEmbedder struct {
	Embedder
	client *api.Client
	model  string
}

func init() {
	ctx := context.Background()
	err := RegisterEmbedder(ctx, "ollama", NewOllamaEmbedder)

	if err != nil {
		panic(err)
	}
}

func NewOllamaEmbedder(ctx context.Context, uri string) (Embedder, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	// I tried to do using plain vanilla native code derived from
	// the logic in github.com/ollama/ollama/server so that in principle
	// you wouldn't need to spin up the ollama server app but there were
	// problems importing vendor deps (what? build tags, maybe?) as
	// well as private variables.

	var cl *api.Client

	if u.Host != "" {

		scheme := "http"

		if q.Has("https") {
			scheme = "https"
		}

		base := &url.URL{
			Scheme: scheme,
			Host:   u.Host,
		}

		http_cl := http.DefaultClient

		cl = api.NewClient(base, http_cl)

	} else {

		v, err := api.ClientFromEnvironment()

		if err != nil {
			return nil, err
		}

		cl = v
	}

	model := q.Get("model")

	e := &OllamaEmbedder{
		client: cl,
		model:  model,
	}

	return e, nil
}

func (e *OllamaEmbedder) Embeddings(ctx context.Context, content string) ([]float64, error) {

	req := &api.EmbeddingRequest{
		Model:  e.model,
		Prompt: content,
	}

	rsp, err := e.client.Embeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return rsp.Embedding, nil
}

// TBD...

func (e *OllamaEmbedder) Embeddings32(ctx context.Context, content string) ([]float32, error) {

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
