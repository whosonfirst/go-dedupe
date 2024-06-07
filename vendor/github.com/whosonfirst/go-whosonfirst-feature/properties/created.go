package properties

import (
	"github.com/tidwall/gjson"
)

func Created(body []byte) int64 {

	rsp := gjson.GetBytes(body, "properties.wof:created")

	if !rsp.Exists() {
		return -1
	}

	return rsp.Int()
}
