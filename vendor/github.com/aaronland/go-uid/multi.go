package uid

import (
	"context"
	"fmt"
	"strings"
)

type MultiUID struct {
	UID
	uids []UID
}

func NewMultiUID(ctx context.Context, uids ...UID) UID {

	r := &MultiUID{
		uids: uids,
	}

	return r
}

func (r *MultiUID) Value() any {
	return r.uids
}

func (r *MultiUID) String() string {

	pairs := make([]string, len(r.uids))

	for idx, uid := range r.uids {

		uid_t := fmt.Sprintf("%T", uid)
		label := strings.Replace(uid_t, "*uid.", "", 1)

		pairs[idx] = fmt.Sprintf("%s#%s", label, uid.String())
	}

	return strings.Join(pairs, " ")
}
