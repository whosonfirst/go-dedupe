package embeddings

import (
	"context"
	"fmt"
	"net/url"

	"github.com/philippgille/chromem-go"
)

type ChromemOllamaEmbedder struct {
	Embedder
	embeddings_func chromem.EmbeddingFunc
}

func init() {
	ctx := context.Background()
	err := RegisterEmbedder(ctx, "chromemollama", NewChromemOllamaEmbedder)

	if err != nil {
		panic(err)
	}
}

func NewChromemOllamaEmbedder(ctx context.Context, uri string) (Embedder, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()
	model := q.Get("model")

	embeddings_func := chromem.NewEmbeddingFuncOllama(model, "")

	e := &ChromemOllamaEmbedder{
		embeddings_func: embeddings_func,
	}

	return e, nil
}

func (e *ChromemOllamaEmbedder) Embeddings(ctx context.Context, content string) ([]float64, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (e *ChromemOllamaEmbedder) Embeddings32(ctx context.Context, content string) ([]float32, error) {
	return e.embeddings_func(ctx, content)
}
