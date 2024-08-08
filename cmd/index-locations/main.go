package main

import (
	"context"
	"log"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	"github.com/whosonfirst/go-dedupe/app/locations/index"
	_ "github.com/whosonfirst/go-dedupe/overture"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
)

func main() {

	ctx := context.Background()
	err := index.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
