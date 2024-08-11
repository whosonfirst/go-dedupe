package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-dedupe/app/whosonfirst/deprecated/migrate"
)

func main() {

	ctx := context.Background()
	err := migrate.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
