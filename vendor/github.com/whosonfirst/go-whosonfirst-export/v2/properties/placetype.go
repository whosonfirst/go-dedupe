package properties

import (
	"errors"

	"github.com/tidwall/gjson"
	_ "github.com/tidwall/sjson"
)

func EnsurePlacetype(feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, "properties.wof:placetype")

	if !rsp.Exists() {
		return feature, errors.New("missing wof:placetype")
	}

	return feature, nil
}
