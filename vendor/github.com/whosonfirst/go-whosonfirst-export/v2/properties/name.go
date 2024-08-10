package properties

import (
	"errors"

	"github.com/tidwall/gjson"
	_ "github.com/tidwall/sjson"
)

func EnsureName(feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, "properties.wof:name")

	if !rsp.Exists() {
		return feature, errors.New("missing wof:name")
	}

	return feature, nil
}
