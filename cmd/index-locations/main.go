package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	_ "github.com/whosonfirst/go-dedupe/overture"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whosonfirst/go-dedupe/app/locations/index"
)

func main() {

	ctx := context.Background()
	err := index.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
