# go-foursquare-places

Go package for working with the Foursquare Places dataset.

## Important

This is work in progress. Documentation and interfaces are incomplete and subject to change.

## Places

Individual Foursquare Place records are represented using the `Place` struct:

```
type Place struct {
	Id            string     `json:"fsq_place_id"`
	Country       string     `json:"country"`
	Address       string     `json:"address"`
	AdminRegion   string     `json:"admin_region"`
	DateClosed    string     `json:"date_closed"`
	DateCreated   string     `json:"date_created"`
	DateRefreshed string     `json:"date_refreshed"`
	Email         string     `json:"email"`
	FacebookId    string     `json:"facebook_id"`
	Instagram     string     `json:"instagram"`
	Latitude      float64    `json:"latitude"`
	Longitude     float64    `json:"longitude"`
	Name          string     `json:"name"`
	PostBox       string     `json:"po_box"`
	PostTown      string     `json:"post_town"`
	PostCode      string     `json:"post_code"`
	Region        string     `json:"region"`
	Telephone     string     `json:"tel"`
	Twitter       string     `json:"twitter"`
	Website       string     `json:"website"`
	Categories    []Category `json:"categories"`
}

### Categories

Categories for individual Foursquare Place records are	represented using the `Category` struct.

```
type Category struct {
	Id     string   `json:"id"`
	Labels []string `json:"labels"`
}
```

## Emitters

The [emitter](emitter) package defines an `Emitter` for iterating through individual Foursquare Places records.

```
type Emitter interface {
	Emit(context.Context) iter.Seq2[*places.Place, error]
	Close() error
}
``

For example:

```
import (
	"context"
	"log/slog"

	"github.com/whosonfirst/go-foursquare-places"
)

func main() {

	ctx := context.Background()
	e, _:= emitter.NewEmitter(ctx, "csv:///path/to/4sq-data.csv.bz2")

	defer e.Close()

	for pl, _ := range e.Emit(ctx) {
		slog.Info("Place", "place", pl)
	}
```

_Error handling omitted for the sake of brevity._

### CSV

At present there is an implementation of the `Emitter` interface for processing bzip2-compressed CSV data exported to a local file using this (DuckDB) query:

```
COPY (SELECT * FROM read_parquet('s3://fsq-os-places-us-east-1/release/dt=2024-11-19/places/parquet/places-*.snappy.parquet')) TO '4sq.csv' (FORMAT CSV, DELIMITER ',', HEADER);
```

CSV emitter constructor URIs take the form of:

```
csv://{PATH_TO_BZIP2_COMPRESSED_CSV_FILE}
```


An important consideration about the CSV emitter is that category data can be missing or incorrect because the source (Parquet) arrays for category IDs and labels are not encoded correctly on export and so it's not really possible to reason about whether a comma (in the CSV output) is a delimeter between records or punctuation in a category label. This will be addressed in future releases.

### DuckDB

It would be simple enough to create a DuckDB emitter using the [go-duckdb](https://github.com/marcboeker/go-duckdb) package. I just haven't done that yet.

## Tools

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/emit cmd/emit/main.go
go build -mod vendor -ldflags="-s -w" -o bin/reverse-geocode cmd/reverse-geocode/main.go
```

### emit

Emit JSON-encoded Foursquare Places records to STDOUT. This tool is mostly for testing the emitter functionality.

```
$> ./bin/emit -h
Usage of ./bin/emit:
  -emitter-uri string
    	A registered /whosonfirst/go-foursquare-places/emitter.Emitter URI.
```

For example:

```
$> ./bin/emit -emitter-uri csv:///usr/local/data/4sq/4sq.csv.bz2
{"fsq_place_id":"4aee4d4688a04abe82d1ace4","country":"ES","address":"Anselm Clavé, 16 Bajo","admin_region":"","date_closed":"","date_created":"2010-05-09","date_refreshed":"2024-06-14","email":"","facebook_id":"","instagram":"","latitude":41.78115297385013,"longitude":3.029574057012216,"name":"Canada House","po_box":"","post_town":"","post_code":"","region":"Gerona","tel":"","twitter":"","website":"http://www.canadahouse.es","categories":[{"id":"4bf58dd8d48988d103951735]","labels":["Retail","Fashion Retail","Clothing Store]"]}]}
{"fsq_place_id":"cb57d89eed29405b908b0b6e","country":"PL","address":"Mickiewicza 8","admin_region":"","date_closed":"","date_created":"2015-05-06","date_refreshed":"2024-06-27","email":"teresa.glowacka.fotos.kolo@neostrada.pl","facebook_id":"","instagram":"","latitude":52.19266718928583,"longitude":18.63343577621856,"name":"Fotos. Zakład fotograficzny. Głowacka T.","po_box":"","post_town":"","post_code":"","region":"Wielkopolskie","tel":"","twitter":"","website":"","categories":[{"id":"4d4b7105d754a06378d81259]","labels":["Retail]"]}]}
{"fsq_place_id":"59a4553d112c6c2b6c378e08","country":"US","address":"","admin_region":"","date_closed":"2019-08-22","date_created":"2017-08-28","date_refreshed":"2024-10-25","email":"","facebook_id":"","instagram":"","latitude":40.774559,"longitude":-73.871849,"name":"CoHo","po_box":"","post_town":"","post_code":"","region":"NY","tel":"","twitter":"","website":"","categories":[{"id":"4bf58dd8d48988d110941735]","labels":["Dining and Drinking","Restaurant","Italian Restaurant]"]}]}
{"fsq_place_id":"4bea3677415e20a110d8e4bb","country":"ID","address":"Bisma75","admin_region":"","date_closed":"","date_created":"2010-05-12","date_refreshed":"2024-07-15","email":"","facebook_id":"","instagram":"","latitude":-6.134261741465453,"longitude":106.86468281476859,"name":"Bisma lounge","po_box":"","post_town":"","post_code":"","region":"Jakarta utara","tel":"","twitter":"","website":"","categories":[{"id":"","labels":[""]}]}
{"fsq_place_id":"dca3aeba404e4b6006620a46","country":"US","address":"18545 Topham St Ste C","admin_region":"","date_closed":"","date_created":"2012-05-21","date_refreshed":"2024-10-12","email":"","facebook_id":"","instagram":"","latitude":34.18099230709591,"longitude":-118.53766860914295,"name":"Grace Motorworks","po_box":"","post_town":"","post_code":"","region":"CA","tel":"","twitter":"","website":"http://gracemotorworks.bzfs.com","categories":[{"id":"52f2ab2ebcbc57f1066b8b44]","labels":["Business and Professional Services","Automotive Service","Automotive Repair Shop]"]}]}
... and so on
```

### reverse-geocode

Reverse-geocode Foursquare Places records against Who's On First records and emit the results as a comma-separated list containing `foursquare_id, whosonfirst_parent_id, whosonfirst_belongs_to_ids` to STDOUT.

```
$> ./bin/reverse-geocode -h
Usage of ./bin/reverse-geocode:
  -emitter-uri string
    	A registered whosonfirst/go-foursquare-places/emitter.Emitter URI.
  -spatial-database-uri string
    	A registered whosonfirst/go-whosonfirst-spatial/database/SpatialDatabase URI to use for perforning reverse geocoding tasks.
  -workers int
    	The maximum number of workers to process reverse geocoding tasks. (default 100)
```

For example:

```
$> ./bin/reverse-geocode \
	-emitter-uri csv:///usr/local/data/4sq/4sq.csv.bz2 \
	-spatial-database-uri 'pmtiles://...' \
	-workers 100 \
	> /usr/local/data/4sq/4sq-wof.csv
```

Which would produce data list this:

```
$> cat /usr/local/data/4sq/4sq-wof.csv
4sq:id,wof:parent_id,wof:belongs_to
5ac2c051898bdc11850c3fce,1762721417,"85672817,102191569,85632429,102031307,1108741693"
4d899d437139b1f7a502b9d4,85902783,"102191569,1159397279,85632509,1159397243"
1da2248395a84f39e8cc81d7,1126095623,"102191581,1125409915,85633735,101752093,85687411"
4f3fd443e4b0b5869fd606b0,101748671,"85682513,102191581,85633111,1377689397,404227561,102063759"
4df118b645dddcd92cb87464,1108940075,"102191581,1377684829,85633111,101748841,1847524019,102064053,85682505"
636d232d8a34153c1022a136,101956319,"85682041,102191577,404567563,85633009,1511777411,102062351"
57bd5be5498e99415bca300b,85930585,"85667983,102191569,85632573,421174961,1376833603"
4e567aaa52b12ddfc1e18326,102073411,"102191569,85632203,85671983"
ef7d80b8e93940a489d2eecb,1729444305,"102191575,85633793,102080987,404484737,85688481"
5139b302e4b0de6b8504573b,102025093,"102191569,85632293,85678743,890466143"
50755c76e4b07ef70e9f4c2a,1158796831,"102191581,1158801659,404474351,85633337,101839323,136253051,85687035"
5ac5f00fa22db7745b9ac792,404538837,"85681497,102191583,85632793,1376953283,136253039,102049207"
91ea56566a4f45bb3bc03802,85940979,"85688697,102191575,404496823,85633793,102082853"
4cba5c124c60a093b9ec45ca,85979145,"85688543,102191575,404521473,85633793,102082371"
...and so on
```

Note: The details of the `-spatial-database-uri` flag are outside the scope of this document. Please consult [whosonfirst/go-whosonfirst-spatial](https://github.com/whosonfirst/go-whosonfirst-spatial?tab=readme-ov-file#database-implementations) for details.

## See also

* https://docs.foursquare.com/data-products/docs/places-os-data-schema
* https://docs.foursquare.com/data-products/docs/access-fsq-os-places
* https://github.com/whosonfirst/go-whosonfirst-spatial