# go-dedupe

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