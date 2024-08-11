# go-dedupe

Go package for resolving duplicate "place" (or venue) locations.

## Important

1. This code was written by and for the Who's On First project but many of the tools are data source (or provider) agnostic.

2. None of this code is especially "fast". It preferences (relative) ease of use and reproducability in favour of speed and other optimizations. Suggestions and gently "clue bats" are welcome.

## Concepts

This code works around (1) common struct and (5) interfaces, and their provider-specific implementations. They are:

* `location.Location` – A Go language struct containing a normalized representation of a place or venue.
* `iterator.Iterator` – A Go language interface for iterating through arbirtrary database sources and emiting JSON-encoded GeoJSON records.
* `location.Parser` – A Go language interface for parsing JSON-encoded GeoJSON records and producing `location.Location` instances.
* `location.Database` – A Go language interface for storing and querying `location.Location` records in a database.
* `embeddings.Embedder` – A Go language interface for generating vector embeddings from input text.
* `vector.Database` – A Go language ...

## Example

### Prune deprecated records

```
$> go run cmd/migrate-deprecated-records/main.go \
	-source-repo /usr/local/data/whosonfirst-data-venue-us-ny \
	-target-repo /usr/local/data/whosonfirst-data-deprecated-venue/
```

### Build locations database

```
$> go run cmd/index-locations/main.go \
	-iterator-uri whosonfirst:// \
	-location-parser-uri whosonfirstvenues:// \
	-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db' \
	/usr/local/data/whosonfirst-data-venue-us-ny/
```

### Compare records

```
$> go run cmd/compare-locations/main.go \
	-source-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db' \
	-target-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db' \
	-workers 50 \
	> /usr/local/data/wof-wof-ny.csv
```

### Process (and deprecate) duplicate records

### Prune deprecated records (again)

## Notes

Given 7.3M Overture places and containerized single-node OpenSearch instance (24GB) on an M-series laptop, storing dense vectors (768) for both name and address fields:

* ~24 hours to index everything with `cmd/index-overture-places`
* 177GB data (OpenSearch)

Querying anything (for example `cmd/compare-alltheplaces`) is brutally slow, like "20771 records in 3h20m0". Logs are full of stuff like:

```
2024-06-14 11:02:46 [2024-06-14T18:02:46,610][INFO ][o.o.a.c.HourlyCron       ] [27612b934c0f] Hourly maintenance succeeds
2024-06-14 11:02:46 [2024-06-14T18:02:46,793][INFO ][o.o.s.l.LogTypeService   ] [27612b934c0f] Loaded [23] customLogType docs successfully!
2024-06-14 11:02:47 [2024-06-14T18:02:47,521][INFO ][o.o.s.i.DetectorIndexManagementService] [27612b934c0f] info deleteOldIndices
2024-06-14 11:02:47 [2024-06-14T18:02:47,525][INFO ][o.o.s.i.DetectorIndexManagementService] [27612b934c0f] info deleteOldIndices
2024-06-14 11:02:47 [2024-06-14T18:02:47,526][INFO ][o.o.s.i.DetectorIndexManagementService] [27612b934c0f] No Old Finding Indices to delete
2024-06-14 11:02:47 [2024-06-14T18:02:47,527][INFO ][o.o.s.i.DetectorIndexManagementService] [27612b934c0f] No Old Alert Indices to delete
2024-06-14 11:02:58 [2024-06-14T18:02:58,648][INFO ][o.o.j.s.JobSweeper       ] [27612b934c0f] Running full sweep
2024-06-14 11:03:05 [2024-06-14T18:03:05,932][INFO ][o.o.s.s.c.FlintStreamingJobHouseKeeperTask] [27612b934c0f] Starting housekeeping task for auto refresh streaming jobs.
2024-06-14 11:03:05 [2024-06-14T18:03:05,983][INFO ][o.o.s.s.c.FlintStreamingJobHouseKeeperTask] [27612b934c0f] Finished housekeeping task for auto refresh streaming jobs.
2024-06-14 11:03:35 [2024-06-14T18:03:35,196][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
2024-06-14 11:05:35 [2024-06-14T18:05:35,235][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
2024-06-14 11:07:35 [2024-06-14T18:07:35,253][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
2024-06-14 11:07:58 [2024-06-14T18:07:58,664][INFO ][o.o.j.s.JobSweeper       ] [27612b934c0f] Running full sweep
2024-06-14 11:09:35 [2024-06-14T18:09:35,274][INFO ][o.o.k.i.KNNCircuitBreaker] [27612b934c0f] [KNN] knn.circuit_breaker.triggered stays set. Nodes at max cache capacity: O9UKyPOTRjWI7rmXV-Z2kg.
```

The (containerized) CPU is pegged at 100% using a steady 15GB of RAM. This is using a single synchronous worker to do lookups. Anything more seems to cause the container to kill itself after a while.

## See also

* https://opensearch.org/docs/latest/search-plugins/knn/settings/
* [Nilesh Dalvi, Marian Olteanu, Manish Raghavan, and Philip Bohannon. 2014. Deduplicating a places database.](https://web.archive.org/web/20160829110541id_/http://wwwconference.org/proceedings/www2014/proceedings/p409.pdf)
* [Carl Yang, Do Huy Hoang, Tomas Mikolov and Jiawei Han. 2019. Place Deduplication with Embeddings](https://arxiv.org/abs/1910.04861)
* [Learning geospatially aware place embeddings via weak-supervision](https://dl.acm.org/doi/10.1145/3557915.3561016)