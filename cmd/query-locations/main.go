package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/whosonfirst/go-dedupe/location"
)

func main() {

	var database_uri string
	var id string
	var geohash string

	flag.StringVar(&database_uri, "database-uri", "", "...")
	flag.StringVar(&id, "id", "", "...")
	flag.StringVar(&geohash, "geohash", "", "...")

	flag.Parse()

	ctx := context.Background()

	db, err := location.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	defer db.Close(ctx)

	if id != "" {

		loc, err := db.GetById(ctx, id)

		if err != nil {
			log.Fatalf("Failed to get ID, %v", err)
		}

		fmt.Println(loc)

	} else {

		cb := func(ctx context.Context, loc *location.Location) error {
			fmt.Println(loc)
			return nil
		}

		err := db.GetWithGeohash(ctx, geohash, cb)

		if err != nil {
			log.Fatalf("Failed to get ID, %v", err)
		}

	}

}
