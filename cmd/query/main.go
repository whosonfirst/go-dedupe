package main

import (
	"context"
	"flag"
	"log"

	"github.com/whosonfirst/go-dedupe/database"
)

func main() {

	var database_uri string
	var query string

	flag.StringVar(&database_uri, "database-uri", "chromem://venues/usr/local/data/venues-dedupe.db?model=mxbai-embed-large", "...")
	flag.StringVar(&query, "query", "", "...")

	flag.Parse()

	ctx := context.Background()

	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	rsp, err := db.Query(ctx, query)

	if err != nil {
		log.Fatalf("Failed to query database, %v", err)
	}

	for _, r := range rsp {
		log.Println(r.ID, r.Content, r.Similarity)
	}
}
