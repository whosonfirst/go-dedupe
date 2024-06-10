package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-dedupe"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	"github.com/whosonfirst/go-dedupe/parser"
)

func main() {

	var parser_uri string

	flag.StringVar(&parser_uri, "parser-uri", "alltheplaces://", "...")

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	prsr, err := parser.NewParser(ctx, parser_uri)

	if err != nil {
		log.Fatalf("Failed to create new parser, %v", err)
	}

	for _, path := range uris {

		r, err := os.Open(path)

		if err != nil {
			log.Fatalf("Failed to open %s for reading, %v", path, err)
		}

		defer r.Close()

		body, err := io.ReadAll(r)

		if err != nil {
			log.Fatalf("Failed to read %s, %v", path, err)
		}

		fc, err := geojson.UnmarshalFeatureCollection(body)

		if err != nil {
			log.Printf("Failed to unmarshal %s, %v\n", path, err)
			continue
		}

		for idx, f := range fc.Features {

			f_body, err := f.MarshalJSON()

			if err != nil {
				log.Fatalf("Failed to marshal feature at offset %d for %s, %v", idx, path, err)
			}

			c, err := prsr.Parse(ctx, f_body)

			if dedupe.IsInvalidRecordError(err) {
				continue
			} else if err != nil {
				log.Fatalf("Failed to parse feature at offset %d for %s, %v", idx, path, err)
			}

			log.Println(c)
		}

	}

}
