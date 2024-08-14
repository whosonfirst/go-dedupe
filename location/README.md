# Locations

## location.Location

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

## location.Parser

```
// Parser is an interface for derive `Location` records from JSON-encoded GeoJSON features.
type Parser interface {
	// Parse derives a `Location` record from a []byte array containing a JSON-encoded GeoJSON feature.
	Parse(context.Context, []byte) (*Location, error)
}
```

### Implementations

#### alltheplaces.AllThePlacesParser

The syntax for creating a new `AllThePlacesParser` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/alltheplaces"
)

ctx := context.Background()
parser, _ := location.NewParser(ctx, "alltheplaces://")
```

#### ilms.ILMSParser

The syntax for creating a new `ILMSParser` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/ilms"
)

ctx := context.Background()
parser, _ := location.NewParser(ctx, "ilms://")
```

#### overture.OverturePlaceParser

The syntax for creating a new `OverturePlaceParser` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/overture"
)

ctx := context.Background()
parser, _ := location.NewParser(ctx, "overtureplaces://")
```

#### whosonfirst.WhosOnFirstVenueParser

The syntax for creating a new `OverturePlaceParser` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	_ "github.com/whosonfirst/go-dedupe/whosonfirst"
)

ctx := context.Background()
parser, _ := location.NewParser(ctx, "whosonfirstvenues://")
```

## location.Database

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

_Note: It is likely that this interface will change to replace the "with callback" methods with methods that return `iter.Seq2` instances._

### Implementations

#### BleveDatabase

The `BleveDatabase` implementation uses the [Bleve indexing library](https://github.com/blevesearch/bleve) to store and query location records.

The syntax for creating a new `BleveDatabase` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
)

ctx := context.Background()
db, _ := location.NewDatabase(ctx, "bleve://{PATH_TO_DATABASE}")
```

Where `{PATH_TO_DATABASE}` is a valid path on the local disk where the Bleve database should be stored.

#### SQLDatabase

The `SQLDatabase` implentation uses the native `database/sql` package to store and query location records.

The syntax for creating a new `SQLDatabase` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/location"
	// Your database/sql driver here
)

ctx := context.Background()
db, _ := location.NewDatabase(ctx, "sql://{ENGINE}?{PARAMETERS}")
```

Where `{ENGINE}` is a valid [database/sql database driver (engine)](https://pkg.go.dev/database/sql).

Valid parameters for the `SQLDatabase` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| dsn| string | yes | A valid valid [database/sql DSN string](https://pkg.go.dev/database/sql) specific to the database driver/engine being used. |
| max-conns | int | no | If defined, sets the maximum number of open connections to the database. |