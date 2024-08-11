# go-dedupe

Go package for resolving duplicate "place" (or venue) locations.

## Documentation

Documentation is incomplete at this time.

## Important

1. This code was written by and for the Who's On First project but many of the tools are data source (or provider) agnostic.
2. None of this code is especially "fast". It preferences (relative) ease of use and reproducability in favour of speed and other optimizations. Suggestions and gently "clue bats" are welcome.
3. This package contains a number of different implementations for a variety of data and storage providers. This reflects the ongoing investigatory nature of the code. At some point in the future some of these implementations may be moved in their own packages or removed entirely.

## Concepts

This code works around (1) common struct and (5) interfaces, and their provider-specific implementations. They are:

* `location.Location` – A Go language struct containing a normalized representation of a place or venue.

* `iterator.Iterator` – A Go language interface for iterating through arbirtrary database sources and emiting JSON-encoded GeoJSON records.
* `location.Parser` – A Go language interface for parsing JSON-encoded GeoJSON records and producing `location.Location` instances.
* `location.Database` – A Go language interface for storing and querying `location.Location` records.
* `embeddings.Embedder` – A Go language interface for generating vector embeddings from input text.
* `vector.Database` – A Go language interface for storing and querying vector embeddings.

The basic working model is as follows:

1. Given a data source or provider, iterate through its records generating and storing `location.Location` records.
2. Given two databases of `location.Location` records:
2a. Derive the set of unique 5-character geohashes from the records in the first database.
2b. For each of those geohashes, find all the `location.Location` records in the second database which a matching geohash and index each record in a vector database.
3. Before a `location.Location` record is stored in a vector database its vector embeddings are derived using an `embeddings.Embedder` instance.
4. Query each of the records matching a given geohash against the records in the vector database; as with the records in the second database, embeddings for each record in the first database are derived using an `embeddings.Embedder` instance.
5. Matching records are emitted as CSV-encoded rows.

For a concrete example, have a look at the code in [app/locations/index](app/locations/index) and [app/locations/compare](app/locations/compare).

### location.Location

```
// Location defines a common format for locations for the purposes of deduplication
type Location struct {
	// The unique ID for the location which is expected to take the form of "{SOURCE_PREFIX}:id={UNIQUE ID}"
	ID string `json:"id"`
	// The name of the location
	Name string `json:"name"`
	// The complete address of the location
	Address string `json:"address"`
	// The principal centroid for the location
	Centroid *orb.Point `json:"centroid"`
	// An arbitrary dictionary of custom metadata properties for the locations. There are a short list of
	// reserved metadata keys which can be queried using the `ReservedMetadataKeys()` or `IsReservedMetadataKey(k)`
	// methods.
	Custom map[string]string `json:"custom,omitempty"`
}
```

### iterator.Iterator

```
// Iterator is an interface for procesing arbitrary data sources that yield individual JSON-encoded GeoJSON Features.
type Iterator interface {
	// Waiting on Go 1.2.3
	// Iterate(context.Context, ...string) iter.Seq2[*geojson.Feature, error]
	IterateWithCallback(context.Context, IteratorCallback, ...string) error
	// Close performs and terminating functions required by the iterator
	Close(context.Context) error
}
```

#### Implementations

##### alltheplaces.AllThePlacesIterator

##### overture.OvertureIterator

##### whosonfirst.WhosOnFirstIterator

### location.Parser

```
// Parser is an interface for derive `Location` records from JSON-encoded GeoJSON features.
type Parser interface {
	// Parse derives a `Location` record from a []byte array containing a JSON-encoded GeoJSON feature.
	Parse(context.Context, []byte) (*Location, error)
}
```

#### Implementations

##### alltheplaces.AllThePlacesParser

##### overture.OverturePlaceParser

##### whosonfirst.WhosOnFirstVenueParser

### location.Database

```
// Database is an interface for storing and querying `Location` records.
type Database interface {
	// AddLocation adds a `Location` record to the underlying database implementation.
	AddLocation(context.Context, *Location) error
	// GetById returns a `Location` record matching an identifier in the underlying database implementation.
	GetById(context.Context, string) (*Location, error)
	// GetGeohashes returns the unique set of geohashes for all the `Location` records stored in the underlying database implementation.
	GetGeohashes(context.Context, GetGeohashesCallback) error
	// GetWithGeohash returns all the `Location` records matching a given geohash in the underlying database implementation.
	GetWithGeohash(context.Context, string, GetWithGeohashCallback) error
	// Close performs and terminating functions required by the database.	
	Close(context.Context) error
}
```

#### Implementations

##### BleveDatabase

##### SQLDatabase

### embeddings.Embedder

```
// Embedder defines an interface for generating (vector) embeddings
type Embedder interface {
	// Embeddings returns the embeddings for a string as a list of float64 values.
	Embeddings(context.Context, string) ([]float64, error)
	// Embeddings32 returns the embeddings for a string as a list of float32 values.	
	Embeddings32(context.Context, string) ([]float32, error)
}
```

#### Implementations

##### ChromemOllamaEmbedder

##### OllamaEmbedder

### vector.Database

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

#### Implementations

##### BleveDatabase

##### ChromemDatabase

##### OpensearchDatabase

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

##### SQLiteDatabase

The `SQLiteDatabase` implementation uses Alex Garcia's [sqlite-vec extension](https://alexgarcia.xyz/blog/2024/sqlite-vec-stable-release/index.html) (and its [Go language bindings](https://alexgarcia.xyz/sqlite-vec/go.html)) to store and query vector embeddings.

## Example

### Prune deprecated records

First, start by moving records flagged as deprecated in the `whosonfirst-data-deprecated-venue` repository. This isn't entirely necessary as the `whosonfirst://` iterator (below) defaults to excluding records which are marked as deprecated.

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

### Compare records (against one another)

```
$> go run cmd/compare-locations/main.go \
	-source-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db' \
	-target-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db' \
	-workers 50 \
	> /usr/local/data/wof-wof-ny.csv

...time passes

2024/08/11 12:30:33 INFO Match geohash=dr5qc threshold=4 similarity=2.7216150760650635 query="T&T Pest Control, 165 10 St Staten Island NY 10306" candidate="T and T Pest Control, 165 10th Street Staten Island NY 10306"
processed 0/5778 records in 41m0.000306167s (started 2024-08-11 11:50:22.939419 -0700 PDT m=+0.643548126)
2024/08/11 12:31:54 INFO Match geohash=dr5rr threshold=4 similarity=3.6093151569366455 query="La Villa Pizzeria & Restrnt, 8207 153rd Avenue Howard Beach NY 11414" candidate="La VIlla Pizzeria, 82-07 153rd Ave. Howard Beach NY 11414"

...and so on
```

The `/usr/local/data/wof-wof-ny.csv` file will look something like this:

```
$> tail -f /usr/local/data/wof-wof-ny.csv
dr5rr,wof:id=353594351,wof:id=353593911,"Cogliano Angelo Jr, 9407 101st Ave Ozone Park NY 11416","Cogliano Angelo Acctnt Jr, 9407 101st Avenue Ozone Park NY 11416",3.018408
dr5xg,wof:id=572126199,wof:id=287214377,"Prosthodontic Associates PC, 1 Hollow Ln Ste 202 New Hyde Park NY 11042","Prosthodontic Associates, 1 Hollow Ln New Hyde Park NY 11042",3.716114
dr5x6,wof:id=303812969,wof:id=269602859,"Hudson Shipping Lines Corp, 20 W Lincoln Ave Valley Stream NY 11580","Hudson Shipping Lines Corp, 20 E Lincoln Ave Valley Stream NY 11580",0.795845
dr7b3,wof:id=370248145,wof:id=253556813,"Pisciotta Capital, 775 Park Dr Huntington Station NY 11793","Pisciotta Capital, 775 Park Ave Huntington NY 11743",3.776641
dr8v9,wof:id=387002999,wof:id=320123265,"Gray Cpa Pc, 16 E Main St Ste 400 Rochester NY 14614","Gray CPA PC, 16 Main St W Rochester NY 14614",2.519037
dr5xq,wof:id=353801261,wof:id=270152357,"Maurice Fur Designer, 69 Merrick Ave Merrick NY 11566","Maurice Fur Designer-Merrick, 69 Merrick Rd North Merrick NY 11566",3.880814
dr5xq,wof:id=555197305,wof:id=253237525,"Matteo's Cafe, 412 Bedford Ave Bellmore NY 11710","Matteos Cafe, 416 Bedford Ave Bellmore NY 11710",3.053007

... and so on
```


### Process (and deprecate) duplicate records

```
...
```

### Prune deprecated records (again)

Move newly deprecated records in the `whosonfirst-data-deprecated-venue` repository.

```
$> go run cmd/migrate-deprecated-records/main.go \
	-source-repo /usr/local/data/whosonfirst-data-venue-us-ny \
	-target-repo /usr/local/data/whosonfirst-data-deprecated-venue/
```

### Compare records (against Overture places)

```
...

```

### Apply concordances (between Who's On First venues and Overture places)

```
...
```

## See also

* https://opensearch.org/docs/latest/search-plugins/knn/settings/
* [Nilesh Dalvi, Marian Olteanu, Manish Raghavan, and Philip Bohannon. 2014. Deduplicating a places database.](https://web.archive.org/web/20160829110541id_/http://wwwconference.org/proceedings/www2014/proceedings/p409.pdf)
* [Carl Yang, Do Huy Hoang, Tomas Mikolov and Jiawei Han. 2019. Place Deduplication with Embeddings](https://arxiv.org/abs/1910.04861)
* [Learning geospatially aware place embeddings via weak-supervision](https://dl.acm.org/doi/10.1145/3557915.3561016)