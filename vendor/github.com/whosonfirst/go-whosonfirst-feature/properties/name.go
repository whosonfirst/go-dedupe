package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

func Name(body []byte) (string, error) {

	rsp := gjson.GetBytes(body, "properties.wof:name")

	if !rsp.Exists() {
		return "", fmt.Errorf("Missing wof:name property")
	}

	name := rsp.String()
	return name, nil
}
