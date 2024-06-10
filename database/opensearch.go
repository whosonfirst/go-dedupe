package database

// https://opensearch.org/docs/latest/search-plugins/semantic-search/
// https://opensearch.org/docs/latest/field-types/supported-field-types/knn-vector/

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	// opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/whosonfirst/go-whosonfirst-opensearch/client"
)

type OpensearchDatabase struct {
	Database
	client *opensearch.Client
	index  string
}

func init() {
	ctx := context.Background()
	err := RegisterDatabase(ctx, "opensearch", NewOpensearchDatabase)

	if err != nil {
		panic(err)
	}
}

func NewOpensearchDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	dsn := q.Get("dsn")

	if dsn == "" {
		return nil, fmt.Errorf("Missing ?dsn= parameter")
	}

	cl, err := client.NewClient(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create opensearch client, %w", err)
	}

	dsn_u, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse dsn (%s), %w", dsn, err)
	}

	index := dsn_u.Path
	index = strings.TrimLeft(index, "/")

	if index == "" {
		return nil, fmt.Errorf("dsn is missing ?index= parameter, '%s'", dsn)
	}

	db := &OpensearchDatabase{
		client: cl,
		index:  index,
	}

	return db, nil
}

func (db *OpensearchDatabase) Add(ctx context.Context, id string, text string, metadata map[string]string) error {
	return nil
}

func (db *OpensearchDatabase) Query(ctx context.Context, text string) ([]*QueryResult, error) {

	results := make([]*QueryResult, 0)
	return results, nil
}
