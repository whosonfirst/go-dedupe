//go:build ollama

package embeddings

import (
	"context"
	"testing"

	"github.com/whosonfirst/go-dedupe"
)

func TestOllamafileEmbeddings(t *testing.T) {
	t.Skip()
}

func TestOllamafileImageEmbeddings(t *testing.T) {

	ctx := context.Background()

	emb, err := NewEmbedder(ctx, "ollama://")

	if err != nil {
		t.Fatalf("Failed to create embedder, %v", err)
	}

	_, err = emb.ImageEmbeddings(ctx, "")

	if !dedupe.IsNotImplementedError(err) {
		t.Fatalf("Unexpected error, %v", err)
	}
}
