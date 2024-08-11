package main

/*

> go run cmd/index-locations/main.go -iterator-uri whosonfirst:// -location-parser-uri whosonfirstvenues:// -location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ca.db' /usr/local/data/whosonfirst-data-venue-us-ca/

*/

import (
	"context"
	"log"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	_ "github.com/whosonfirst/go-dedupe/overture"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"

	"github.com/whosonfirst/go-dedupe/app/locations/index"
)

func main() {

	ctx := context.Background()
	err := index.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
