package main

import (
	"context"
	"flag"
	"log"

	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-dedupe/location"
	"github.com/whosonfirst/go-dedupe/vector"
)

func main() {

	var vector_database_uri string
	var name string
	var address string

	var kv_pairs multi.KeyValueString

	flag.StringVar(&vector_database_uri, "vector-database-uri", "chromem://venues/usr/local/data/venues-dedupe.db?model=mxbai-embed-large", "...")
	flag.StringVar(&name, "name", "", "...")
	flag.StringVar(&address, "address", "", "...")
	flag.Var(&kv_pairs, "metadata", "...")

	flag.Parse()

	ctx := context.Background()

	db, err := vector.NewDatabase(ctx, vector_database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	loc := &location.Location{
		Name:    name,
		Address: address,
	}

	if len(kv_pairs) > 0 {

		metadata := make(map[string]string, 0)

		for _, kv := range kv_pairs {
			metadata[kv.Key()] = kv.Value().(string)
		}

		loc.Custom = metadata
	}

	rsp, err := db.Query(ctx, loc)

	if err != nil {
		log.Fatalf("Failed to query database, %v", err)
	}

	for _, r := range rsp {
		log.Println(r.ID, r.Content, r.Similarity)
	}
}
