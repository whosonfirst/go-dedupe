package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

func Id(body []byte) (int64, error) {

	rsp := gjson.GetBytes(body, "properties.wof:id")

	if !rsp.Exists() {
		return 0, fmt.Errorf("Missing wof:id property")
	}

	id := rsp.Int()

	if id < 0 {
		return 0, fmt.Errorf("Invalid or unrecognized ID value (%d)", id)
	}

	return id, nil
}
