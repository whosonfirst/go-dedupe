# Command line tools

```
$> make cli
cd ../ && make cli && cd -
go build -tags sqlite,sqlite_vec,duckdb,ollama -mod vendor -ldflags="-s -w" -o bin/compare-locations cmd/compare-locations/main.go
go build -tags sqlite,sqlite_vec,duckdb,ollama -mod vendor -ldflags="-s -w" -o bin/index-locations cmd/index-locations/main.go
go build -mod vendor -ldflags="-s -w" -o bin/wof-assign-concordances cmd/wof-assign-concordances/main.go
go build -mod vendor -ldflags="-s -w" -o bin/wof-migrate-deprecated cmd/wof-migrate-deprecated/main.go
go build -mod vendor -ldflags="-s -w" -o bin/wof-process-duplicates cmd/wof-process-duplicates/main.go
/usr/local/whosonfirst/go-dedupe/cmd
```

## Build tags

Location and vector database implementations are enabled through the use of [Go build tags](https://pkg.go.dev/go/build). The default Makefile `cli` target enables the following tags automatically: `sqlite,sqlite_vec,duckdb,ollama`.

### chromem

Required to use the `ChromemDatabase` or the `ChromemOllamaEmbedder` implementations of the `vector.Database` and `embeddings.Embedder` interfaces, respectively, using the [philippgille/chromem-go](https://github.com/philippgille/chromem-go) package.

### duckdb

Required to use the `DuckDBDatabase` implementations of the `vector.Database` and `location.Database` interfaces using the [marcboeker/go-duckdb](https://github.com/marcboeker/go-duckdb) package.

_If you encounter problems building the tools it might have something to do with the way `go-duckdb` is vendored. The best place to start debugging things is [this section in the go-duckdb documentation](https://github.com/marcboeker/go-duckdb?tab=readme-ov-file#vendoring)._

### ollama

Required to use the `OllamaEmbedder` implementation of the `embeddings.Embedder` interface, using the [Ollama REST API](https://github.com/ollama/ollama/blob/main/docs/api.md).

### opensearch

Required to use the `OpensearchDatabase` implementation of the `vector.Database` interface, using the [opensearch-project/opensearch-go](https://github.com/opensearch-project/opensearch-go) package.

### sqlite

Required to use the `SQLDatabase` implementation of the `location.Database` interface using the [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) package.

### sqlite_vec

Required to use the `SQLiteDatabase` implementation of the `vector.Database` interface using the [asg017/sqlite-vec-go-bindings](https://github.com/asg017/sqlite-vec-go-bindings) package.

## Tools

### compare-locations

Compare two location databases and emit matching records as CSV-encoded rows.

```
$> ./bin/compare-locations -h
Compare two location databases and emit matching records as CSV-encoded rows.
Usage:
	 ./bin/compare-locations [options]
Valid options are:
  -monitor-uri string
    	A valid sfomuseum/go-timings.Monitor URI. (default "counter://PT60S")
  -source-location-database-uri string
    	A valid whosonfirst/go-dedupe/location.Database URI.
  -target-location-database-uri string
    	A valid whosonfirst/go-dedupe/location.Database URI.
  -threshold float
    	The threshold value for matching records. Whether this value is greater than or lesser than a matching value will be dependent on the vector database in use. (default 4)
  -vector-database-dsn string
    	A valid whosonfirst/go-dedupe/vector.Database DSN string. If the parameter contains the string "{geohash}" then that string will be replaced, at runtime, with the value of the geohash being compared. This will have the effect of creating a vector database per geohash. This value will be used to replace any "{vector-database-dsn}" strings in the -vector-database-uri flag. (default "{tmp}{geohash}.db?cache=shared&mode=memory")
  -vector-database-embedder-uri string
    	A valid whosonfirst/go-dedupe/embeddings.Embedder URI. This value will be used to replace any "{vector-database-embedder-uri}" strings in the -vector-database-uri flag. (default "ollama://?model={vector-database-model}")
  -vector-database-model string
    	The name of the model to use comparing records in the location database against records in the vector database. This value will be used to replace any "{vector-database-model}" strings in the -vector-database-uri and -vector-database-embedder-uri flags. (default "mxbai-embed-large")
  -vector-database-uri string
    	A valid whosonfirst/go-dedupe/vector.Database URI. (default "sqlite://?model={vector-database-model}&dsn={vector-database-dsn}&embedder-uri={vector-database-embedder-uri}&max-distance=4&max-results=10&dimensions=1024&compression=none")
  -verbose
    	Enable verbose (debug) logging.
  -workers int
    	The number of simultaneous worker processes to use. (default 10)
```

For example:

```
$> ./bin/compare-locations \
	-source-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db' \
	-target-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db' \
	-workers 50 \
	> /usr/local/data/wof-wof-ny.csv
```

### index-locations

Populate (index) a location database from data/provider source..

```
$> ./bin/index-locations -h
Populate (index) a location database from data/provider source..
Usage:
	 ./bin/index-locations [options] uri(N) uri(N)
 Valid options are:
  -iterator-uri string
    	A valid whosonfirst/go-dedupe/iterator.Iterator URI.
  -location-database-uri string
    	A valid whosonfirst/go-dedupe/location.Database URI.
  -location-parser-uri string
    	A valid whosonfirst/go-dedupe/location.Parser URI.
  -monitor-uri string
    	A valid sfomuseum/go-timings.Monitor URI. (default "counter://PT60S")
  -verbose
    	Enable verbose (debug) logging.
```

For example:

```
$> ./bin/index-locations \
	-iterator-uri whosonfirst:// \
	-location-parser-uri whosonfirstvenues:// \
	-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db&max-conns=1' \
	/usr/local/data/whosonfirst-data-venue-us-ny/
```

### wof-assign-concordances

Assign concordances from a data/provider source to a Who's On First repository..

```
$> ./bin/wof-assign-concordances -h
Assign concordances from a data/provider source to a Who's On First repository..
Usage:
	 ./bin/wof-assign-concordances [options] uri(N) uri(N)
Valid options are:
  -concordance-as-int
    	If true cast the concordance ID as an int64
  -concordance-namespace string
    	The namespace of the concordance being applied.
  -concordance-predicate string
    	The predicate of the concordance being applies. (default "id")
  -mark-is-current
    	If true the addition of a cocordance will mark this record as mz:is_current=1
  -reader-uri string
    	A valid whosonfirst/go-reader.Reader URI for reading WOF records from.
  -verbose
    	Enable verbose (debug) logging.
  -whosonfirst-label string
    	The "label" used to identify WOF records. Valid options are: source, target. (default "target")
  -writer-uri string
    	A valid whosonfirst/go-writer.Writer URI for writing WOF records to.
```

For example:

```
$> ./bin/wof-assign-concordances \
	-reader-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	-writer-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	-concordance-namespace ovtr \
	-concordance-predicate id \
	/usr/local/data/ovtr-wof-ny.csv
```

### wof-migrate-deprecated

Migrate deprecated records from one Who's On First repository to another.

```
$> ./bin/wof-migrate-deprecated -h
Migrate deprecated records from one Who's On First repository to another.
Usage:
	 ./bin/wof-migrate-deprecated [options]
Valid options are:
  -source-repo string
    	The path to the Who's On First repository that deprecated records will be removed from.
  -target-repo string
    	The path to the Who's On First repository that deprecated records will be added from.
  -verbose
    	Enable verbose (debug) logging.
```

For example:

```
$> ./bin/wof-migrate-deprecated \
	-source-repo /usr/local/data/whosonfirst-data-venue-us-ny \
	-target-repo /usr/local/data/whosonfirst-data-deprecated-venue/
```

### wof-process-duplicates

Process duplicate records in a Who's On First repository (which means deprecate and mark as superseding or superseded by where necessary).

```
$> ./bin/wof-process-duplicates -h
Process duplicate records in a Who's On First repository (which means deprecate and mark as superseding or superseded by where necessary).
Usage:
	 ./bin/wof-process-duplicates [options] uri(N) uri(N)
Valid options are:
  -reader-uri string
    	A valid whosonfirst/go-reader.Reader URI that records to be processed will be read from.
  -verbose
    	Enable verbose (debug) logging.
  -writer-uri string
    	A valid whosonfirst/go-writer.Writer URI where updated records will be written to. (default "stdout://")
```

For example:

```
$> ./bin/wof-process-duplicates \
	-reader-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	-writer-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	/usr/local/data/wof-wof-ny.csv
```