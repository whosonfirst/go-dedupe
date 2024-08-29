package embeddings

import (
	"context"
)

// NullEmbedder implements the `Embedder` interface using an Null API endpoint to derive embeddings.
type NullEmbedder struct {
	Embedder
}

func init() {
	ctx := context.Background()
	err := RegisterEmbedder(ctx, "null", NewNullEmbedder)

	if err != nil {
		panic(err)
	}
}

func NewNullEmbedder(ctx context.Context, uri string) (Embedder, error) {

	e := &NullEmbedder{}
	return e, nil
}

func (e *NullEmbedder) Embeddings(ctx context.Context, content string) ([]float64, error) {

	embeddings := make([]float64, 0)
	return embeddings, nil
}

func (e *NullEmbedder) Embeddings32(ctx context.Context, content string) ([]float32, error) {

	e64, err := e.Embeddings(ctx, content)

	if err != nil {
		return nil, err
	}

	return asFloat32(e64), nil
}

func (e *NullEmbedder) ImageEmbeddings(ctx context.Context, data []byte) ([]float64, error) {

	embeddings := make([]float64, 0)
	return embeddings, nil
}

func (e *NullEmbedder) ImageEmbeddings32(ctx context.Context, data []byte) ([]float32, error) {

	e64, err := e.ImageEmbeddings(ctx, data)

	if err != nil {
		return nil, err
	}

	return asFloat32(e64), nil
}
