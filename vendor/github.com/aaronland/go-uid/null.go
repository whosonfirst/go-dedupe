package uid

import (
	"context"
	"fmt"
	"log"
)

// NULL_SCHEME is the URI scheme used to identify NullProvider instances.
const NULL_SCHEME string = "null"

func init() {
	ctx := context.Background()
	RegisterProvider(ctx, NULL_SCHEME, NewNullProvider)
}

// type NullProvider implements the Provider interface to return empty (null) identifiers.
type NullProvider struct {
	Provider
}

// type NullProvider implements the UID interface to return empty (null) identifiers.
type NullUID struct {
	UID
}

// NewNullProvider
//
//	null://
func NewNullProvider(ctx context.Context, uri string) (Provider, error) {
	pr := &NullProvider{}
	return pr, nil
}

func (n *NullProvider) UID(ctx context.Context, args ...interface{}) (UID, error) {
	return NewNullUID(ctx)
}

func (n *NullProvider) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}

func NewNullUID(ctx context.Context) (UID, error) {
	n := &NullUID{}
	return n, nil
}

func (n *NullUID) Value() any {
	return ""
}

func (n *NullUID) String() string {
	return fmt.Sprintf("%v", n.Value())
}
