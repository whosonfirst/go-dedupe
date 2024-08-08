package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	_ "github.com/whosonfirst/go-dedupe/overture"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
	_ "gocloud.dev/blob/fileblob"
	_ "github.com/mattn/go-sqlite3"	
	"github.com/whosonfirst/go-dedupe/app/locations/compare"
)

func main() {

	ctx := context.Background()
	err := compare.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
