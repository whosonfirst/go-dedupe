package main

import (
	"context"
	"flag"
	"log"

	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-dedupe/database"
)

func main() {

	var database_uri string
	var query string

	var kv_pairs multi.KeyValueString

	flag.StringVar(&database_uri, "database-uri", "chromem://venues/usr/local/data/venues-dedupe.db?model=mxbai-embed-large", "...")
	flag.StringVar(&query, "query", "", "...")
	flag.Var(&kv_pairs, "metadata", "...")

	flag.Parse()

	ctx := context.Background()

	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	metadata := make(map[string]string, 0)

	for _, kv := range kv_pairs {
		metadata[kv.Key()] = kv.Value().(string)
	}

	rsp, err := db.Query(ctx, query, metadata)

	if err != nil {
		log.Fatalf("Failed to query database, %v", err)
	}

	for _, r := range rsp {
		log.Println(r.ID, r.Content, r.Similarity)
	}
}
