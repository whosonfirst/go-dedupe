# go-brooklynintegers-api

Go package for the Brooklyn Integers API.

## Documentation

Documentation is incomplete at this time.

## Usage

## Simple

```
package main

import (
	"fmt"
	"github.com/aaronland/go-brooklynintegers-api"
)

func main() {

	client := api.NewAPIClient()
	i, _ := client.CreateInteger()

	fmt.Println(i)
}
```

## Less simple

```
import (
       "fmt"
       "github.com/aaronland/go-brooklynintegers-api"
)

func main() {

	client := api.NewAPIClient()

	params := url.Values{}
	method := "brooklyn.integers.create"

	rsp, _ := client.ExecuteMethod(method, &params)
	i, _ := rsp.Int()

	fmt.Println(i)
}
```

## Tools

### int

Mint one or more Brooklyn Integers.

```
$> ./bin/int -h
Usage of ./bin/int:
  -count int
    	The number of Brooklyn Integers to mint (default 1)
```

### proxy-server

This tool has been moved to the [go-brooklynintegers-proxy](https://github.com/aaronland/go-brooklynintegers-proxy#proxy-server) package.

## See also

* http://brooklynintegers.com/
* http://brooklynintegers.com/api
* https://github.com/aaronland/go-brooklynintegers-proxy
