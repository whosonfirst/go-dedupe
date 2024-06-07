package properties

import (
	"github.com/tidwall/gjson"
)

func LastModified(body []byte) int64 {

	rsp := gjson.GetBytes(body, "properties.wof:lastmodified")

	if !rsp.Exists() {
		return -1
	}

	return rsp.Int()
}
