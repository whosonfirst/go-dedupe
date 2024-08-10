package uid

import (
	"context"
	"fmt"
	"log"
	"time"
)

const YMD_SCHEME string = "ymd"

func init() {
	ctx := context.Background()
	RegisterProvider(ctx, YMD_SCHEME, NewYMDProvider)
}

type YMDProvider struct {
	Provider
}

type YMDUID struct {
	UID
	date time.Time
}

func NewYMDProvider(ctx context.Context, uri string) (Provider, error) {
	pr := &YMDProvider{}
	return pr, nil
}

func (pr *YMDProvider) UID(ctx context.Context, args ...interface{}) (UID, error) {

	date := time.Now()

	if len(args) == 1 {

		str_date := args[0].(string)

		t, err := time.Parse("20060102", str_date)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse time %s, %w", str_date, err)
		}

		date = t
	}

	return NewYMDUID(ctx, date)
}

func (pr *YMDProvider) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}

func NewYMDUID(ctx context.Context, date time.Time) (UID, error) {

	u := &YMDUID{
		date: date,
	}

	return u, nil
}

func (u *YMDUID) Value() any {
	return u.date.Format("20060102")
}

func (u *YMDUID) String() string {
	return fmt.Sprintf("%v", u.Value())
}
