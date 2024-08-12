# Vector databases

## vector.Database

```
// Database defines an interface for adding and querying vector embeddings of `location.Location` records.
type Database interface {
	// Add adds a `Location` record to the underlying database implementation.	
	Add(context.Context, *location.Location) error
	// Query results a list of `QueryResult` instances for records matching a `location.Location` in the underlying database implementation.
	Query(context.Context, *location.Location) ([]*QueryResult, error)
	// MeetsThreshold returns a boolean value indicating whether a `QueryResult` instance satisfies a given threshold value.
	MeetsThreshold(context.Context, *QueryResult, float64) (bool, error)
	// Close performs and terminating functions required by the database.		
	Close(context.Context) error
}
```

### Implementations

#### BleveDatabase

#### ChromemDatabase

#### OpensearchDatabase

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

#### SQLiteDatabase

The `SQLiteDatabase` implementation uses Alex Garcia's [sqlite-vec extension](https://alexgarcia.xyz/blog/2024/sqlite-vec-stable-release/index.html) (and its [Go language bindings](https://alexgarcia.xyz/sqlite-vec/go.html)) to store and query vector embeddings.

The syntax for creating a new `SQLiteDatabase` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/vector"
	_ "github.com/mattn/go-sqlite3"
)

ctx := context.Background()
, _ := vector.NewDatabase(ctx, "sqlite://?{PARAMETERS")
```

Valid parameters for the `SQLiteDatabase` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| dsn| string | yes | DSN strings are discussed below. |
| embedder-uri | string | yes | A valid `Embedder` URI. |
| dimensions | int | no | The dimensionality of the vector embeddings to store and query. Default is `768`. |
| max_distance | float | no | The maximum distance between any two records being queried. Default is `5.0` |
| max_results | int | no | The maximum number of results to return for any given query. Default is `10` |
| compression | string | no | The type of compression to use when storing (and querying) embeddings. Valid options are: none, quantize, matroyshka. Default is `none`. Consult the [sqlite-vec extension](https://alexgarcia.xyz/blog/2024/sqlite-vec-stable-release/index.html) documentation for details. |
| refresh | bool | no | A boolean flag to indicate whether existing records should be updated. Default is `false`. |
| max-conns | int | no | If defined, sets the maximum number of open connections to the database. |

By default DSN strings take the form detailed in the [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) documentation.

If a DSN contains the string `{tmp}` then the (SQLiteDatabase) code will create a new SQLite database to be used for storing and querying documents. That database will be created in whatever temporary folder the operating system defines and removed the (SQLiteDatabase) `Close` method is invoked.
