package emitter

import (
	"context"
)

func init() {
	ctx := context.Background()
	RegisterEmitter(ctx, "null", NewNullEmitter)
}

// NullEmitter implements the `Emitter` interface for appearing to crawl records but not doing anything.
type NullEmitter struct {
	Emitter
}

// NewNullEmitter() returns a new `NullEmitter` instance configured by 'uri' in the form of:
//
//	null://
func NewNullEmitter(ctx context.Context, uri string) (Emitter, error) {

	idx := &NullEmitter{}
	return idx, nil
}

// WalkURI() does nothing.
func (idx *NullEmitter) WalkURI(ctx context.Context, index_cb EmitterCallbackFunc, uri string) error {
	return nil
}
