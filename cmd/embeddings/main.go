package main

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/whosonfirst/go-dedupe/embeddings"
)

func main() {

	var embedder_uri string

	flag.StringVar(&embedder_uri, "embedder-uri", "ollama://?model=mxbai-embed-large", "...")

	flag.Parse()

	content := strings.Join(flag.Args(), " ")

	ctx := context.Background()

	em, err := embeddings.NewEmbedder(ctx, embedder_uri)

	if err != nil {
		log.Fatal(err)
	}

	e, err := em.Embeddings(ctx, content)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(len(e))
}
