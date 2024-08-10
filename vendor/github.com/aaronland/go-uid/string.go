package uid

import (
	"context"
	"fmt"
	"log"
	"net/url"
)

const STRING_SCHEME string = "string"

func init() {
	ctx := context.Background()
	RegisterProvider(ctx, STRING_SCHEME, NewStringProvider)
}

type StringProvider struct {
	Provider
	string string
}

type StringUID struct {
	UID
	string string
}

func NewStringProvider(ctx context.Context, uri string) (Provider, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse string, %w", err)
	}

	q := u.Query()
	s := q.Get("string")

	if s == "" {
		return nil, fmt.Errorf("Empty string")
	}

	pr := &StringProvider{
		string: s,
	}

	return pr, nil
}

func (pr *StringProvider) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}

func (pr *StringProvider) UID(ctx context.Context, args ...interface{}) (UID, error) {
	return NewStringUID(ctx, pr.string)
}

func NewStringUID(ctx context.Context, s string) (UID, error) {

	u := StringUID{
		string: s,
	}

	return &u, nil
}

func (u *StringUID) Value() any {
	return u.string
}

func (u *StringUID) String() string {
	return fmt.Sprintf("%v", u.Value())
}
