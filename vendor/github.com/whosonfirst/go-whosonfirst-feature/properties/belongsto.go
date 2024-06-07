package properties

import (
	"github.com/tidwall/gjson"
)

func BelongsTo(body []byte) []int64 {

	by := make([]int64, 0)

	rsp := gjson.GetBytes(body, "properties.wof:belongsto")

	for _, r := range rsp.Array() {
		by = append(by, r.Int())
	}

	return by
}
