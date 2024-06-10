package main

import (
	"context"
	"flag"
	"log"

	"github.com/whosonfirst/go-dedupe/database"
)

func main() {

	var database_uri string

	flag.StringVar(&database_uri, "database-uri", "sqlite://?dsn=test.db", "...")

	flag.Parse()

	ctx := context.Background()

	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	texts := []string{
		"Hello world",
		"omg",
		"wtf",
		"bbq",
	}

	for _, t := range texts {
		db.Add(ctx, "#", t, nil)
	}

	queries := []string{
		"hello there",
		"hello world",
		"bar-b-que",
		"Hello world",
	}

	for _, q := range queries {

		log.Println("Q", q)

		qr, err := db.Query(ctx, q)

		if err != nil {
			log.Println("Q", q, err)
			continue
		}

		for i, r := range qr {

			log.Println("QR", i, q, r)
		}

	}

}
