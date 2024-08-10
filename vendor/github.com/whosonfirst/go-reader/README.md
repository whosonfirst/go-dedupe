# go-reader

There are many interfaces for reading files. This one is ours. It returns `io.ReadSeekCloser` instances.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-reader.svg)](https://pkg.go.dev/github.com/whosonfirst/go-reader)

### Example

Readers are instantiated with the `reader.NewReader` method which takes as its arguments a `context.Context` instance and a URI string. The URI's scheme represents the type of reader it implements and the remaining (URI) properties are used by that reader type to instantiate itself.

For example to read files from a directory on the local filesystem you would write:

```
package main

import (
	"context"
	"github.com/whosonfirst/go-reader"
	"io"
	"os"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "file:///usr/local/data")
	fh, _ := r.Read(ctx, "example.txt")
	defer fh.Close()
	io.Copy(os.Stdout, fh)
}
```

There is also a handy "null" reader in case you need a "pretend" reader that doesn't actually do anything:

```
package main

import (
	"context"
	"github.com/whosonfirst/go-reader"
	"io"
	"os"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "null://")
	fh, _ := r.Read(ctx, "example.txt")
	defer fh.Close()
	io.Copy(os.Stdout, fh)
}
```

## Interfaces

### reader.Reader

```
type Reader interface {
	Read(context.Context, string) (io.ReadSeekCloser, error)
	ReaderURI(context.Context, string) string
}
```

## Custom readers

Custom readers need to:

1. Implement the interface above.
2. Announce their availability using the `go-reader.RegisterReader` method on initialization, passing in an initialization function implementing the `go-reader.ReaderInitializationFunc` interface.

For example, this is how the [go-reader-http](https://github.com/whosonfirst/go-reader-http) reader is implemented:

```
package reader

import (
	"context"
	"errors"
	wof_reader "github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-ioutil"
	"io"
	_ "log"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type HTTPReader struct {
	wof_reader.Reader
	url      *url.URL
	throttle <-chan time.Time
}

func init() {

	ctx := context.Background()

	schemes := []string{
		"http",
		"https",
	}

	for _, s := range schemes {

		err := wof_reader.RegisterReader(ctx, s, NewHTTPReader)

		if err != nil {
			panic(err)
		}
	}
}

func NewHTTPReader(ctx context.Context, uri string) (wof_reader.Reader, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	rate := time.Second / 3
	throttle := time.Tick(rate)

	r := HTTPReader{
		throttle: throttle,
		url:      u,
	}

	return &r, nil
}

func (r *HTTPReader) Read(ctx context.Context, uri string) (io.ReadSeekCloser, error) {

	<-r.throttle

	u, _ := url.Parse(r.url.String())
	u.Path = filepath.Join(u.Path, uri)

	url := u.String()

	rsp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != 200 {
		return nil, errors.New(rsp.Status)
	}

	fh, err := ioutil.NewReadSeekCloser(rsp.Body)

	if err != nil {
		return nil, err
	}

	return fh, nil
}

func (r *HTTPReader) ReaderURI(ctx context.Context, uri string) string {
	return uri
}
```

And then to use it you would do this:

```
package main

import (
	"context"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"	
	"io"
	"os"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "https://data.whosonfirst.org")
	fh, _ := r.Read(ctx, "101/736/545/101736545.geojson")
	defer fh.Close()
	io.Copy(os.Stdout, fh)
}
```

## Available readers

### "blob"

Read files from any registered [Go Cloud](https://gocloud.dev/howto/blob/) `Blob` source. For example:

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-blob"
	_ "gocloud.dev/blob/s3blob"	
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "s3://whosonfirst-data?region=us-west-1")
}
```

* https://github.com/whosonfirst/go-reader-blob

### github://

Read files from a GitHub repository.

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-github"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "github://{GITHUB_OWNER}/{GITHUB_REPO}")

	// to specify a specific branch you would do this:
	// r, _ := reader.NewReader(ctx, "github://{GITHUB_OWNER}/{GITHUB_REPO}?branch={GITHUB_BRANCH}")
}
```

* https://github.com/whosonfirst/go-reader-github

### githubapi://

Read files from a GitHub repository using the GitHub API.

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-github"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "githubapi://{GITHUB_OWNER}/{GITHUB_REPO}?access_token={GITHUBAPI_ACCESS_TOKEN}")

	// to specify a specific branch you would do this:
	// r, _ := reader.NewReader(ctx, "githubapi://{GITHUB_OWNER}/{GITHUB_REPO}/{GITHUB_BRANCH}?access_token={GITHUBAPI_ACCESS_TOKEN}")
}
```

* https://github.com/whosonfirst/go-reader-github

### http:// and https://

Read files from an HTTP(S) endpoint.

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "https://{HTTP_HOST_AND_PATH}")
}
```

* https://github.com/whosonfirst/go-reader-http

### file://

Read files from a local filesystem.

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "file://{PATH_TO_DIRECTORY}")
}
```

If you are importing the `go-reader-blob` package and using the GoCloud's [fileblob](https://gocloud.dev/howto/blob/#local) driver then instantiating the `file://` scheme will fail since it will have already been registered. You can work around this by using the `fs://` scheme. For example:

```
r, _ := reader.NewReader(ctx, "fs://{PATH_TO_DIRECTORY}")
```

* https://github.com/whosonfirst/go-reader

### null://

Pretend to read files.

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "null://")
}
```

### repo://

This is a convenience scheme for working with Who's On First data repositories.

It will update a URI by appending a `data` directory to its path and changing its scheme to `fs://` before invoking `reader.NewReader` with the updated URI.

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "repo:///usr/local/data/whosonfirst-data-admin-ca")
}
```

### stdin://

Read "files" from `STDIN`

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "stdin://")
}
```

And then to use, something like:

```
> cat README.md | ./bin/read -reader-uri stdin:// - | wc -l
     339
```

Note the use of `-` for a URI. This is the convention (when reading from STDIN) but it can be whatever you want it to be.

## See also

* https://github.com/whosonfirst/go-writer
