# go-dedupe

Go package for resolving duplicate "place" (or venue) locations.

## Documentation

Documentation, in particular the `godoc` documentation,  is incomplete at this time.

## Background

For background please have a look at the [ De-duplicating Who's On First venues with vector embeddings ](https://whosonfirst.org/blog/2024/08/16/dedupe/) and, for historical context, the [Who's On First ꞉fist-bump꞉ OpenStreetMap](https://whosonfirst.org/blog/2017/10/24/whosonfirst-sotmus-2017/) blog posts.

## Important

1. This code was written by and for the Who's On First project but many of the tools are data source (or provider) agnostic.
2. This package contains a number of different implementations for a variety of data and storage providers. This reflects the ongoing investigatory nature of the code. At some point in the future some of these implementations may be moved in their own packages or removed entirely.
3. None of this code is especially "fast". It preferences (relative) ease of use and reproducability in favour of speed and other optimizations. It can often hours, sometimes days, to process large datasets. Suggestions and gentle "clue bats" are welcome.

## Concepts

This code works around (1) common struct and (5) interfaces, and their provider-specific implementations. They are:

* [location.Location](location/README.md#locationlocation) – A Go language struct containing a normalized representation of a place or venue.
* [location.Parser](location/README.md#locationparser) – A Go language interface for parsing JSON-encoded GeoJSON records and producing `location.Location` instances.
* [location.Database](location/README.md#locationdatabase) – A Go language interface for storing and querying `location.Location` records.
* [iterator.Iterator](iterator/README.md) – A Go language interface for iterating through arbirtrary database sources and emiting JSON-encoded GeoJSON records.
* [embeddings.Embedder](embeddings/README.md) – A Go language interface for generating vector embeddings from input text.
* [vector.Database](vector/README.md) – A Go language interface for storing and querying vector embeddings.

The basic working model is as follows:

1. Given a data source or provider, iterate through its records generating and storing `location.Location` records.
2. Given two databases of `location.Location` records, one of them the "source" and the other the "target":
3. Derive the set of unique 5-character geohashes from the records in the "target" database.
4. For each of those geohashes, find all the `location.Location` records in the "source" database which a matching geohash and index each record in a vector database.
5. Store each matching ("source") `location.Location` record in a vector database deriving its embeddings using an `embeddings.Embedder` instance.
6. Query each of the ("target") records matching a given geohash against the records in the vector database; as with the records in the second database, embeddings for each record in the first database are derived using an `embeddings.Embedder` instance.
7. Matching records are emitted as CSV-encoded rows.

What happens with those CSV rows of matching records is left for implementors to decide. For a concrete example, have a look at the code in [app/locations/index](app/locations/index), [app/locations/compare](app/locations/compare) and the [compare](compare) package.

There are a few things to note about this approach:

* A 5-character geohash represents an area of approximately 2.4 km. In the future it may be the case that a longer geohash will be stored (in the location database) and a variable length geohash will be queried based on properties that can be derived about a location. For example, a venue in the center of Manhattan might use a longer, more precise geohash, versus a venue in a rural area might use a shorter, more inclusive, geohash.
* Likewise, if `location.Location` records have been supplemented with Who's On First hierarchies (on ingest or at runtime) then they might also be filtered by geohash _and_ region to account for the fact that the same geohash can span multiple administrative boundaries (for example `dr5re`).
* This code works best with small and short-lived (temporary) vector databases on disk or in memory. Storing and querying millions of venue records and their embeddings on consumer grade hardware (my laptop) is generally slow and impractical. Many (but not all, yet) of the `vector.Database` implementations have been configured with the ability to create (and remove) temporary databases automatically. Details are discussed in the [documentation for vector databases](vector/README.md).

As of this writing most of the work has been centered around the SQLite and DuckDB implementations for [location databases](location/README.md) and [vector databases](https://github.com/whosonfirst/go-dedupe/blob/main/vector/README.md) and the Ollama implementation for [generating embeddings](embeddings/README.md#ollamaembedder). Details for each are discussed in their respective packages.

## Data sources (providers)

The following data source (providers) have working implementations (iterators and location parsers) for use with this package:

* [All The Places](https://www.alltheplaces.xyz/)
* [Institute of Museum and Library Services](https://www.imls.gov/research-evaluation/data-collection/museum-data-files) (Museum Data Files)
* [Overture Data](https://docs.overturemaps.org/guides/places/) (Places)
* [Who's On First](https://github.com/whosonfirst-data/?q=whosonfirst-data-venue&type=all&language=&sort=) (Venues)

## Location database implementations

* [Bleve](location#blevedatabase)
* [DuckDB](location#sqldatabase)
* [SQLite](location#sqldatabase)

## Vector database implementation

* [Bleve](vector#blevedatabase)
* [Chromem](vector#chromemdatabase)
* [DuckDB](vector#duckdb)
* [SQLite](vector#sqlitedatabase)

## Embeddings implementations

* [Chromem (Ollama API)](embeddings#chromemollamaembedder)
* [llamafile (API)](embeddings#llamafileembedder)
* [Ollama (API)](embeddings#ollamaembedder)
* [OpenCLIP (API)](embeddings#openclipembedder)

### Embeddings implementations (third-party)

* [MLX CLIP](https://github.com/sfomuseum/go-embeddings-mlxclip)

## Example

```
$> make cli
go build -tags sqlite,sqlite_vec,duckdb,ollama -mod vendor -ldflags="-s -w" -o bin/compare-locations cmd/compare-locations/main.go
go build -tags sqlite,sqlite_vec,duckdb,ollama -mod vendor -ldflags="-s -w" -o bin/index-locations cmd/index-locations/main.go
go build -mod vendor -ldflags="-s -w" -o bin/wof-assign-concordances cmd/wof-assign-concordances/main.go
go build -mod vendor -ldflags="-s -w" -o bin/wof-migrate-deprecated cmd/wof-migrate-deprecated/main.go
go build -mod vendor -ldflags="-s -w" -o bin/wof-process-duplicates cmd/wof-process-duplicates/main.go
```

Documentation and details for these tools (and their build tags) can be found in the [cmd/README.md](cmd/README.md) file.

The following examples will show the work flow for prepping and comparing a Who's On First (WOF) venue repository against a collection of Overture Data place records and then assigning concordances (matching records) to the WOF venues.

### Prune deprecated records

First, start by moving records flagged as deprecated in the `whosonfirst-data-deprecated-venue` repository. This isn't entirely necessary as the `whosonfirst://` iterator (below) defaults to excluding records which are marked as deprecated.

```
$> ./bin/wof-migrate-deprecated \
	-source-repo /usr/local/data/whosonfirst-data-venue-us-ny \
	-target-repo /usr/local/data/whosonfirst-data-deprecated-venue/
```

### Build locations database(s)

First create a locations database for all these Who's On First venues.

```
$> ./bin/index-locations \
	-iterator-uri whosonfirst:// \
	-location-parser-uri whosonfirstvenues:// \
	-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny.db&max-conns=1' \
	/usr/local/data/whosonfirst-data-venue-us-ny/
```

Next create a locations database for Overture Data places, in this example venues with a confidence level of 0.95 or higher. For details on how to create a file like please consult the [documentation for the `whosonfirst/go-overture package.](https://github.com/whosonfirst/go-overture?tab=readme-ov-file#exporting-overture-parquet-files-to-line-separated-json) 

```
$> ./bin/index-locations \
	-iterator-uri overture:// \
	-location-parser-uri overtureplaces:// \
	-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/overture-locations.db&max-conns=1' \
	/usr/local/data/overture/venues-0.95.geojsonl.bz2
```

### Compare records (against one another)

Now compare the Who's On First locations database against itself. This step is performed in order to deduplicate WOF records in the same database (repo).

```
$> ./bin/compare-locations \
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

Process any duplicate records in the Who's On First repository. This will mark records as deprecated, superseding or superseded by as necessary.

```
$> ./bin/wof-process-duplicates \
	-reader-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	-writer-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	/usr/local/data/wof-wof-ny.csv
```

### Prune deprecated records (again)

Move newly deprecated records in the `whosonfirst-data-deprecated-venue` repository. This isn't strictly necessary but because there are so many venues it's nice to move known-bad records in to their own isolated sandbox.

```
$> ./bin/wof-migrate-deprecated \
	-source-repo /usr/local/data/whosonfirst-data-venue-us-ny \
	-target-repo /usr/local/data/whosonfirst-data-deprecated-venue/
```

Rebuild the Who's On First location database:

```
$> ./bin/index-locations \
	-iterator-uri whosonfirst:// \
	-location-parser-uri whosonfirstvenues:// \
	-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny-2.db' \
	/usr/local/data/whosonfirst-data-venue-us-ny/
```

### Compare Who's On First records against Overture records

Compare Who's On First records against Overture records, writing matching records to a CSV file on disk.

```
$> ./bin/compare-locations \
	-source-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/overture-locations.db' \
	-target-location-database-uri 'sql://sqlite3?dsn=/usr/local/data/whosonfirst-ny-2.db' \
	-workers 50 \
	> /usr/local/data/overture/ovtr-wof-ny.csv

...time passes (a lot of it)
```

Eventually this process will complete and looking at the resultant CSV file there are over 25, 000 matching records between these two data sources:

```
$> wc -l /usr/local/data/ovtr-wof-ny.csv 
   25538 /usr/local/data/ovtr-wof-ny.csv
```

_At some point in the future it may be possible to specify different "targets" where matching records are emitted to (both locally and remotely) but for the time being a CSV file will do._

### Apply concordances (between Who's On First venues and Overture places)

Update the Who's On First records with 
```
$> ./bin/wof-assign-concordances \
	-reader-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	-writer-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	-concordance-namespace ovtr \
	-concordance-predicate id \
	/usr/local/data/ovtr-wof-ny.csv
```

_The `wof-assign-concordances` tool a tool to allow you to explictly mark matching Who's On First records as being "current". As of this writing that is a manual process. In the future there may be code/logic to assign this property (`mz:is_current`) based on one or more "confidence" levels define by data providers but those details have not been finalized yet._

## See also

* https://opensearch.org/docs/latest/search-plugins/knn/settings/
* [Nilesh Dalvi, Marian Olteanu, Manish Raghavan, and Philip Bohannon. 2014. Deduplicating a places database.](https://web.archive.org/web/20160829110541id_/http://wwwconference.org/proceedings/www2014/proceedings/p409.pdf)
* [Carl Yang, Do Huy Hoang, Tomas Mikolov and Jiawei Han. 2019. Place Deduplication with Embeddings](https://arxiv.org/abs/1910.04861)
* [Learning geospatially aware place embeddings via weak-supervision](https://dl.acm.org/doi/10.1145/3557915.3561016)