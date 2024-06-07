package main

import (
	"context"
	"flag"
	"log"

	"github.com/whosonfirst/go-dedupe/database"
)

func main() {

	var database_uri string

	flag.StringVar(&database_uri, "database-uri", "chromem://venues/usr/local/data/venues-dedupe.db?model=mxbai-embed-large", "...")

	flag.Parse()

	ctx := context.Background()

	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database")
	}

	err = db.Add(ctx, "1", "Hello world", nil)

	if err != nil {
		log.Fatalf("Failed to add record")
	}
}
