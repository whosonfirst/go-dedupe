# Iterators

## iterator.Iterator

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

### Implementations

#### alltheplaces.AllThePlacesIterator

The `AllThePlacesIterator` processes one or more [All The Places](https://www.alltheplaces.xyz/) GeoJSON FeatureCollection files. For example:

```
$> go run cmd/index-locations/main.go \
	-location-database-uri null:// \
	-location-parser-uri alltheplaces:// \
	-iterator-uri alltheplaces:// \
	/usr/local/data/alltheplaces/*.geojson
```

The syntax for creating a new `AllThePlacesIterator` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
)

ctx := context.Background()
loc, _ := location.NewLocation(ctx, "alltheplaces://")
```

#### ilms.ILMSIterator

The `ILMSIterator` processes one or more records in the ILMS [Museum Data Files](https://www.imls.gov/research-evaluation/data-collection/museum-data-files) CSV records. For example:

```
$> go run cmd/index-locations/main.go \
	-location-database-uri null:// \
	-location-parser-uri ilms:// \
	-iterator-uri ilms:// \
	/usr/local/data/2018_csv_museum_data_files/MuseumFile2018_File*.csv
```

The syntax for creating a new `ILMSIterator` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/ilms"
)

ctx := context.Background()
loc, _ := location.NewLocation(ctx, "ilms://")
```

#### overture.OvertureIterator

The `OvertureIterator` processes one or more JSON-L files (optionally bzip-compressed) containing Overture Data GeoJSON Feature records. For example:

```
$> go run cmd/index-locations/main.go \
	-location-database-uri null:// \
	-location-parser-uri overtureplaces:// \
	-iterator-uri 'overture://?bucket-uri=file:///' \
	/usr/local/data/overture/places-geojson/venues-0.95.geojsonl.bz2
```

The syntax for creating a new `OvertureIterator` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/overture"
)

ctx := context.Background()
loc, _ := location.NewLocation(ctx, "overture://?{PARAMETERS}")
```

Valid parameters for the `OvertureIterator` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| bucket-uri | A valid `gocloud.dev/blob` URI | yes | Default is `file:///` |

#### whosonfirst.WhosOnFirstIterator

The `WhosOnFirstIterator` processes one or more GeoJSON features returned by an underlying [whosonfirst/go-whosonfirst-iterate/v2](https://github.com/whosonfirst/go-whosonfirst-iterate) instance. For example:

```
$> go run cmd/index-locations/main.go \
	-location-database-uri null:// \
	-location-parser-uri whosonfirst:// \
	-iterator-uri 'whosonfirst://' \
	/usr/local/data/whosonfirst-data-venue-us-ca
```

The syntax for creating a new `WhosOnFirstIterator` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
)

ctx := context.Background()
loc, _ := location.NewLocation(ctx, "whosonfirst://?{PARAMETERS}")
```

Valid parameters for the `WhosOnFirstIterator` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| iterator-uri | A valid `whosonfirst/go-whosonfirst-iterate/v2` URI | yes | Default is `repo://?exclude=properties.edtf:deprecated=.*` |
