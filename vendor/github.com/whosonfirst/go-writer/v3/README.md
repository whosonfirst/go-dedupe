# go-whosonfirst-writer

Go package that provides a common interface for writing data to multiple sources.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-writer.svg)](https://pkg.go.dev/github.com/whosonfirst/go-writer)

## Example

Writers are instantiated with the `writer.NewWriter` method which takes as its arguments a `context.Context` instance and a URI string. The URI's scheme represents the type of writer it implements and the remaining (URI) properties are used by that writer type to instantiate itself. For example:

```
import (
       "context"
       "github.com/whosonfirst/go-writer/v2"
       "strings"
)

func main() {

	ctx := context.Background()
	
	r := strings.NewReader("Hello world")
	
	wr, _ := writer.NewWriter(ctx, "fs:///usr/local/example")
	wr.Write(ctx, wr, "hello/world.txt", r)
}
```

_Error handling omitted for the sake of brevity._

## Writers

The following writers are exported by this package. Additional writers are defined in separate [go-writer-*](https://github.com/whosonfirst/?q=go-writer-&type=all&language=&sort=) packages.

### cwd://

Implements the `Writer` interface for writing documents to the current working directory.

### fs://

Implements the `Writer` interface for writing documents as files on a local disk.

### io://

Implements the `Writer` interface for writing documents to an `io.Writer` instance.

### multi://

Implements the `Writer` interface for writing documents to multiple `Writer` instances.

### null://

Implements the `Writer` interface for writing documents to nowhere.

### repo://

Implements the `Writer` interface for writing documents to a Who's On First "data" directory.

### stdout://

Implements the `Writer` interface for writing documents to STDOUT.

## See also

* https://github.com/whosonfirst/?q=go-writer-&type=all&language=&sort=
* https://github.com/whosonfirst/go-reader
