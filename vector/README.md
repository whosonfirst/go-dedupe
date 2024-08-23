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

_tl;dr – As of this writing most of the work and testing (and successes) has been happening around the [SQLiteDatabase and DuckDB](#sqlitedatabase) implementations._

#### BleveDatabase

The `BleveDatabase` implementation uses the [Bleve indexing library](https://github.com/blevesearch/bleve) to store and query vector embeddings.

The syntax for creating a new `BleveDatabase` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/vector"
)

ctx := context.Background()
, _ := vector.NewDatabase(ctx, "bleve://{PATH}?{PARAMETERS")
```

Valid parameters for the `BleveDatabase` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| embedder-uri | string | yes | A valid `Embedder` URI. |
| dimensions | int | no | The dimensionality of the vector embeddings to store and query. Default is `768`. |

By default `{PATH}` strings take the form of a local path on disk.

If a path contains the string `{tmp}` then the (BleveDatabase) code will create a new Bleve database to be used for storing and querying documents. That database will be created in whatever temporary folder the operating system defines and removed the (BleveDatabase) `Close` method is invoked.

Note: This code was last tested before the adoption of small, temporary databases. When indexing 7.3M Overture Data place records the final database was both really big (multiple dozens of GB if memory serves) and really slow. It is worth revisiting how effective things are with on-demand per-geohash databases.

Use of the `BleveDatabase` implementation requires tools be built with the `-bleve` tag.

#### ChromemDatabase

The `ChromemDatabase` implementation uses the [philippgille/chromem-go](https://github.com/philippgille/chromem-go) package to store and query vector embeddings. In turn `chromem-go` uses the [Ollama application's REST API](https://github.com/ollama/ollama?tab=readme-ov-file#rest-api) to generate embeddings for a text. This package assumes that the Ollama application has already installed, is running and set up to use the models necessary to generate embeddings. Please consult the [Ollama documentation](https://github.com/ollama/ollama) for details.

The syntax for creating a new `ChromemDatabase` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/vector"
)

ctx := context.Background()
, _ := vector.NewDatabase(ctx, "chromem://?{PARAMETERS")
```

Valid parameters for the `ChromemDatabase` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| model | string| yes | The name of the model you want to Ollama API to use when generating embeddings. |

Note: This code was last tested before the adoption of small, temporary databases. When indexing 7.3M Overture Data place records the final (on-disk) database was both really big (almost 100 GB, I think) and really slow. It is worth revisiting how effective things are with on-demand and in-memory per-geohash databases.

Use of the `ChromemDatabase` implementation requires tools be built with the `-chromem` tag.

#### DuckDB

The `DuckDBDatabase` uses the [DuckDB](https://duckdb.org/) database and the [VSS extension](https://duckdb.org/docs/extensions/vss) to store and query vector embeddings.

The syntax for creating a new `DuckDBDatabase` is:

```
import (
	"context"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/whosonfirst/go-dedupe/vector"
)

ctx := context.Background()
, _ := vector.NewDatabase(ctx, "duckdb://?{PARAMETERS")
```

Valid parameters for the `DuckDBDatabase` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| embedder-uri | string | yes | A valid `Embedder` URI. |
| dimensions | int | no | The dimensionality of the vector embeddings to store and query. Default is `768`. |
| max_distance | float | no | The maximum distance between any two records being queried. Default is `5.0` |
| max_results | int | no | The maximum number of results to return for any given query. Default is `10` |
| refresh | bool | no | A boolean flag to indicate whether existing records should be updated. Default is `false`. |
| max-conns | int | no | If defined, sets the maximum number of open connections to the database. |

`DuckDBDatabase` do not take a DSN parameter since, as of this writing, vector embeddings [are not (can not) be persisted to disk](https://duckdb.org/docs/extensions/vss#persistence) yet.

Use of the `DuckDBDatabase` implementation requires tools be built with the `-duckdb` tag.

#### OpensearchDatabase

The `OpensearchDatabase` uses the [OpenSearch](https://opensearch.org/) document storage engine to store and query vector embeddings.

The syntax for creating a new `OpensearchDatabase` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/vector"
)

ctx := context.Background()
, _ := vector.NewDatabase(ctx, "opensearch://?{PARAMETERS")
```

Valid parameters for the `OpensearchDatabase` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| client-uri | string | yes | A URI string that can be parsed by the [whosonfirst/go-whosonfirst-opensearch/client.ClientOptionsFromURI](https://github.com/whosonfirst/go-whosonfirst-opensearch/blob/main/client/client.go#L40C6-L40C26) method. |
| model | string| yes | The name of the model you want to use when generating embeddings. |

Some things to note:

Given 7.3M Overture places and a containerized single-node OpenSearch instance (24GB) on an M-series laptop, storing dense vectors (768) for both name and address fields indexing required:

* ~24 hours to store everything
* 177GB of disk space (OpenSearch data)

Querying anything (for example `cmd/compare-alltheplaces`) is brutally slow, like "20771 records in 3h20m0" and the log files are full of "knn.circuit_breaker.triggered" errors. The (containerized) CPU was often pegged at 100% using a steady 15GB of RAM. This is using a single synchronous worker to do lookups. Anything more seems to cause the container to kill itself after a while.

Additionally, all of the steps required to [configure Opensearch as a vector database](https://opensearch.org/docs/latest/search-plugins/semantic-search/) are assumed to have happened _before_ constructor (above) is invoked. This code was last tested before the adoption of small, temporary databases and it is something worth revisiting but this will also require adding code to spin up, configure and tear down individual (per-geohash) OpenSearch indices on demand. Have a look at the [Makefile is this directory](Makefile) for an example of all the steps necessary to make this possible.

Use of the `OpenSearchDatabase` implementation requires tools be built with the `-opensearch` tag.

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

Use of the `SQLiteDatabase` implementation requires tools be built with the `-sqlite_vec` tag.