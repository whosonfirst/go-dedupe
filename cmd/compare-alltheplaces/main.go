package main

/*

[asc][asc@SD-931-4][11:05:22] /usr/local/whosonfirst/go-whosonfirst-dedupe                                                                                                                     > go run cmd/compare-alltheplaces/main.go /usr/local/data/alltheplaces/dunkin_us.geojson

*/

import (
	"context"
	"flag"
	"io"
	"log"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-dedupe"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
	// "github.com/whosonfirst/go-dedupe/database"
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/overture"
	"github.com/whosonfirst/go-dedupe/parser"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	var location_database_uri string
	var parser_uri string
	var monitor_uri string

	// var bucket_uri string
	// var is_bzipped bool

	var threshold float64

	// flag.StringVar(&database_uri, "database-uri", "opensearch://?dsn=https%3A%2F%2Flocalhost%3A9200%2Fdedupe%3Fusername%3Dadmin%26password%3DKJHFGDFJGSJfsdkjfhsdoifruwo45978h52dcn%26insecure%3Dtrue%26require-tls%3Dtrue&model=9dgHD5ABSoo-6k3cWDqn&bulk-index=false", "...")
	//flag.StringVar(&database_uri, "database-uri", "chromem://venues/usr/local/data/venues.db?model=mxbai-embed-large", "...")

	flag.StringVar(&location_database_uri, "location-database-uri", "", "...")
	flag.StringVar(&parser_uri, "parser-uri", "alltheplaces://", "...")
	flag.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "...")

	// flag.StringVar(&bucket_uri, "bucket-uri", "file:///", "...")
	// flag.BoolVar(&is_bzipped, "is-bzip2", true, "...")

	flag.Float64Var(&threshold, "threshold", 0.95, "...")
	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	slog.Info("Create database")
	db, err := location.NewDatabase(ctx, location_database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	slog.Info("Create parser")
	prsr, err := parser.NewParser(ctx, parser_uri)

	if err != nil {
		log.Fatalf("Failed to create new parser, %v", err)
	}

	/*
		source_bucket, err := bucket.OpenBucket(ctx, bucket_uri)

		if err != nil {
			log.Fatalf("Failed to open source bucket, %v", err)
		}
	*/

	cmp, err := dedupe.NewComparator(ctx, db, prsr, os.Stdout)

	if err != nil {
		log.Fatalf("Failed to create new comparator, %v", err)
	}

	defer cmp.Flush()

	monitor, err := timings.NewMonitor(ctx, monitor_uri)

	if err != nil {
		log.Fatalf("Failed to create monitor, %v", err)
	}

	monitor.Start(ctx, os.Stderr)
	defer monitor.Stop(ctx)

	// Anything more seems to make a local (Docker) OS instance SAD
	max_workers := 1
	throttle := make(chan bool, max_workers)

	for i := 0; i < max_workers; i++ {
		throttle <- true
	}

	total_matches := int64(0)
	total_features := int64(0)

	wg := new(sync.WaitGroup)

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

		slog.Info("Unmarshal", "path", path)

		fc, err := geojson.UnmarshalFeatureCollection(body)

		if err != nil {
			slog.Warn("Failed to unmarshal feature collection", "path", path, "error", err)
			continue
		}

		features := int64(0)
		matches := int64(0)

		for idx, f := range fc.Features {

			<-throttle

			wg.Add(1)

			go func(f *geojson.Feature) {

				defer func() {
					wg.Done()
					throttle <- true
					monitor.Signal(ctx)
				}()

				atomic.AddInt64(&features, 1)
				atomic.AddInt64(&total_features, 1)

				logger := slog.Default()
				logger = logger.With("path", path)
				logger = logger.With("offset", idx)

				f_body, err := f.MarshalJSON()

				if err != nil {
					logger.Warn("Failed to marshal feature", "error", err)
					return
				}

				is_match, err := cmp.Compare(ctx, f_body, threshold)

				if err != nil {
					slog.Warn("Failed to compare feature", "path", path, "error", err)
					return
				}

				if is_match {
					atomic.AddInt64(&matches, 1)
					atomic.AddInt64(&total_matches, 1)
				}
			}(f)
		}

		slog.Info("Matches", "path", path, "features", atomic.LoadInt64(&features), "matches", atomic.LoadInt64(&matches), "total features", atomic.LoadInt64(&total_features), "total matches", atomic.LoadInt64(&total_matches))

	}

}

/*

	at org.opensearch.search.SearchService$2.lambda$onResponse$0(SearchService.java:592) ~[opensearch-2.14.0.jar:2.14.0]
	at org.opensearch.action.ActionRunnable.lambda$supply$0(ActionRunnable.java:74) ~[opensearch-2.14.0.jar:2.14.0]
	at org.opensearch.action.ActionRunnable$2.doRun(ActionRunnable.java:89) ~[opensearch-2.14.0.jar:2.14.0]
	at org.opensearch.common.util.concurrent.AbstractRunnable.run(AbstractRunnable.java:52) ~[opensearch-2.14.0.jar:2.14.0]
	... 8 more
[2024-06-14T01:45:06,869][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
[2024-06-14T01:47:06,908][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
[2024-06-14T01:47:07,056][INFO ][o.o.j.s.JobSweeper       ] [27612b934c0f] Running full sweep
[2024-06-14T01:47:35,779][WARN ][o.o.m.j.JvmGcMonitorService] [27612b934c0f] [gc][90477] overhead, spent [2.8s] collecting in the last [3.6s]
[2024-06-14T01:49:06,899][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
[2024-06-14T01:51:06,971][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
[2024-06-14T01:52:07,023][INFO ][o.o.s.s.c.FlintStreamingJobHouseKeeperTask] [27612b934c0f] Starting housekeeping task for auto refresh streaming jobs.
[2024-06-14T01:52:07,044][INFO ][o.o.j.s.JobSweeper       ] [27612b934c0f] Running full sweep
[2024-06-14T01:52:07,084][INFO ][o.o.s.s.c.FlintStreamingJobHouseKeeperTask] [27612b934c0f] Finished housekeeping task for auto refresh streaming jobs.
[2024-06-14T01:53:07,058][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
[2024-06-14T01:53:42,346][INFO ][o.o.m.j.JvmGcMonitorService] [27612b934c0f] [gc][90842] overhead, spent [467ms] collecting in the last [1s]
[2024-06-14T01:53:43,579][INFO ][o.o.m.j.JvmGcMonitorService] [27612b934c0f] [gc][90843] overhead, spent [340ms] collecting in the last [1.2s]

*/
