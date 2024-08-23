//go:build llamafile

package embeddings

import (
	"context"
	"testing"
)

func TestLlamafileEmbedder(t *testing.T) {

	ctx := context.Background()

	emb, err := NewEmbedder(ctx, "llamafile://")

	if err != nil {
		t.Fatalf("Failed to create embedder, %v", err)
	}

	rsp, err := emb.Embeddings(ctx, "Hello world")

	if err != nil {
		t.Fatalf("Failed to derive embeddings, %v", err)
	}

	if len(rsp) == 0 {
		t.Fatalf("Empty embedding")
	}
}
