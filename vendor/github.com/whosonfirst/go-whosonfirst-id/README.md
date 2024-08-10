# go-whosonfirst-id

Go package for generating valid Who's On First IDs.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-id.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-id)

## Example

_Error handling omitted for the sake of brevity._

### Simple

```
package main

import (
	"context"
	"fmt"
	"testing"
)

func main() {

	ctx := context.Background()
	pr, _ := NewProvider(ctx)

	id, _ := pr.NewID(ctx)
	fmt.Println(id)
}
```

### Fancy

The default `Provider` does not pre-generate or cache IDs. To do so create a custom `Provider` that does, use the handy `NewProviderWithURI` method:

```
package main

import (
	_ "github.com/aaronland/go-uid-whosonfirst"
)

import (
	"context"
	"fmt"
	_ "github.com/aaronland/go-uid-proxy"	
)

func main() {

	ctx := context.Background()

	uri := "proxy:///?provider=whosonfirst://&minimum=5&pool=memory%3A%2F%2F"
	cl, _ := NewProviderWithURI(ctx, uri)

	id, _ := cl.NextInt(ctx)
	fmt.Println(id)
}
```

This expects a valid [go-uid-proxy](https://github.com/aaronland/go-uid-proxy) URI string.

## See also

* https://github.com/aaronland/go-uid
* https://github.com/aaronland/go-uid-proxy
* https://github.com/aaronland/go-uid-artisanal
* https://github.com/aaronland/go-uid-whosonfirst
* https://github.com/aaronland/go-brooklynintegers-api