package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-dedupe/app/whosonfirst/concordances/assign"
)

func main() {

	ctx := context.Background()
	err := assign.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
