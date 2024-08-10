package uid

import (
	"context"
	"strconv"
)

type Int64UID struct {
	UID
	int64 int64
}

func NewInt64UID(ctx context.Context, i int64) (UID, error) {

	u := Int64UID{
		int64: i,
	}

	return &u, nil
}

func (u *Int64UID) Value() any {
	return u.int64
}

func (u *Int64UID) String() string {
	return strconv.FormatInt(u.int64, 10)
}
