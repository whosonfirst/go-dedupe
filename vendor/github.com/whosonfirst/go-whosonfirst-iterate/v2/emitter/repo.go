package emitter

import (
	"context"
	"fmt"
	"path/filepath"
)

func init() {
	ctx := context.Background()
	RegisterEmitter(ctx, "repo", NewRepoEmitter)
}

// RepoEmitter implements the `Emitter` interface for crawling records in a Who's On First style data directory.
type RepoEmitter struct {
	Emitter
	// emitter is the underlying `DirectoryEmitter` instance for crawling records.
	emitter Emitter
}

// NewDirectoryEmitter() returns a new `RepoEmitter` instance configured by 'uri' in the form of:
//
//	repo://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewRepoEmitter(ctx context.Context, uri string) (Emitter, error) {

	directory_idx, err := NewDirectoryEmitter(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new directory emitter, %w", err)
	}

	idx := &RepoEmitter{
		emitter: directory_idx,
	}

	return idx, nil
}

// WalkURI() appends 'uri' with "data" and then walks that directory and for each file (not excluded by any
// filters specified when `idx` was created) invokes 'index_cb'.
func (idx *RepoEmitter) WalkURI(ctx context.Context, index_cb EmitterCallbackFunc, uri string) error {

	abs_path, err := filepath.Abs(uri)

	if err != nil {
		return fmt.Errorf("Failed to derive absolute path for '%s', %w", uri, err)
	}

	data_path := filepath.Join(abs_path, "data")
	return idx.emitter.WalkURI(ctx, index_cb, data_path)
}
