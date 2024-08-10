package reader

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
)

func init() {

	ctx := context.Background()

	err := RegisterReader(ctx, "repo", NewRepoReader)

	if err != nil {
		panic(err)
	}

}

// NewRepoReader is a convenience method to update 'uri' by appending a `data`
// directory to its path and changing its scheme to `fs://` before invoking
// NewReader with the updated URI.
func NewRepoReader(ctx context.Context, uri string) (Reader, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	root := filepath.Join(u.Path, "data")

	uri = fmt.Sprintf("fs://%s", root)
	return NewReader(ctx, uri)
}
