package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-dedupe/app/whosonfirst/duplicates/process"
)

func main() {

	ctx := context.Background()
	err := process.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
