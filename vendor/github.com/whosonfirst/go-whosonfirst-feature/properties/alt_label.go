package properties

import (
	"github.com/tidwall/gjson"
)

func AltLabel(body []byte) (string, error) {
	rsp := gjson.GetBytes(body, "properties.src:alt_label")
	return rsp.String(), nil
}
