package writer

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
)

func init() {

	ctx := context.Background()

	err := RegisterWriter(ctx, "repo", NewRepoWriter)

	if err != nil {
		panic(err)
	}

}

// NewRepoWriter is a convenience method to update 'uri' by appending a `data`
// directory to its path and changing its scheme to `fs://` before invoking
// NewWriter with the updated URI.
func NewRepoWriter(ctx context.Context, uri string) (Writer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	root := filepath.Join(u.Path, "data")

	uri = fmt.Sprintf("fs://%s", root)
	return NewWriter(ctx, uri)
}
