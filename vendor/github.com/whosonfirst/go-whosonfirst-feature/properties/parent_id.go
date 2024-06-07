package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

// https://github.com/whosonfirst/whosonfirst-properties/tree/main/properties/wof#parent_id
func ParentId(body []byte) (int64, error) {

	rsp := gjson.GetBytes(body, "properties.wof:parent_id")

	if !rsp.Exists() {
		return 0, fmt.Errorf("Missing wof:parent_id property")
	}

	id := rsp.Int()

	// https://github.com/whosonfirst/whosonfirst-properties/tree/main/properties/wof#parent_id

	if id < -4 {
		return 0, fmt.Errorf("Invalid or unrecognized parent ID value (%d)", id)
	}

	return id, nil
}
