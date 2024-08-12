package main

import (
	"context"
	"log"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	_ "github.com/whosonfirst/go-dedupe/ilms"
	_ "github.com/whosonfirst/go-dedupe/overture"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
	_ "gocloud.dev/blob/fileblob"

	"github.com/whosonfirst/go-dedupe/app/locations/compare"
)

func main() {

	ctx := context.Background()
	err := compare.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
