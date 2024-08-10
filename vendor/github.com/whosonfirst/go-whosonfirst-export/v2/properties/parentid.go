package properties

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func EnsureParentId(feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, "properties.wof:parent_id")

	if rsp.Exists() {
		return feature, nil
	}

	return sjson.SetBytes(feature, "properties.wof:parent_id", -1)
}
