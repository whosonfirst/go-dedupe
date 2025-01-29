package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	_ "log/slog"
	"os"

	"github.com/whosonfirst/go-dedupe/embeddings"
)

const stdin string = "-"

func main() {

	var embedder_uri string
	var is_image bool

	flag.StringVar(&embedder_uri, "embedder-uri", "null://", "A registered whosonfirst/go-dedupe/embeddings.Embedder URI.")
	flag.BoolVar(&is_image, "image", false, "A boolean flag indicating whether to derive embeddings for an image.")

	flag.Parse()

	ctx := context.Background()

	emb, err := embeddings.NewEmbedder(ctx, embedder_uri)

	if err != nil {
		log.Fatalf("Failed to create embedder, %v", err)
	}

	for _, path := range flag.Args() {

		var data []float64
		var input []byte
		var err error

		if path == stdin {

			input, err = io.ReadAll(os.Stdin)

			if err != nil {
				log.Fatalf("Failed to read data from STDIN, %v", err)
			}

		} else {

			input, err = os.ReadFile(path)

			if err != nil {
				log.Fatalf("Failed to read data from %s, %v", path, err)
			}

		}

		if is_image {
			data, err = emb.ImageEmbeddings(ctx, input)
		} else {
			data, err = emb.Embeddings(ctx, string(input))
		}

		if err != nil {
			log.Fatalf("Failed to derive embeddings, %v", err)
		}

		enc := json.NewEncoder(os.Stdout)
		err = enc.Encode(data)

		if err != nil {
			log.Fatalf("Failed to encode embeddings, %v", err)
		}
	}

}
