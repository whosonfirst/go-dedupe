package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

func Placetype(body []byte) (string, error) {

	rsp := gjson.GetBytes(body, "properties.wof:placetype")

	if !rsp.Exists() {
		return "", fmt.Errorf("Missing wof:placetype property")
	}

	placetype := rsp.String()

	return placetype, nil
}
