//go:build llamafile

package embeddings

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"testing"
)

func TestLlamafileEmbeddings(t *testing.T) {

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

func TestLlamafileImageEmbeddings(t *testing.T) {

	ctx := context.Background()

	emb, err := NewEmbedder(ctx, "llamafile://")

	if err != nil {
		t.Fatalf("Failed to create embedder, %v", err)
	}

	im_path := "../fixtures/1527845303_walrus.jpg"

	im_r, err := os.Open(im_path)

	if err != nil {
		t.Fatalf("Failed to open %s for reading, %v", im_path, err)
	}

	defer im_r.Close()

	im_body, err := io.ReadAll(im_r)

	if err != nil {
		t.Fatalf("Failed to read data from %s, %v", im_path, err)
	}

	im_b64 := base64.StdEncoding.EncodeToString(im_body)

	rsp, err := emb.ImageEmbeddings(ctx, im_b64)

	if err != nil {
		t.Fatalf("Failed to derive embeddings, %v", err)
	}

	if len(rsp) == 0 {
		t.Fatalf("Empty embedding")
	}
}
