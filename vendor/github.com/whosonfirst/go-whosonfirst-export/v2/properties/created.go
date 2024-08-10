package properties

import (
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func EnsureCreated(feature []byte) ([]byte, error) {

	var err error

	now := int32(time.Now().Unix())

	created := gjson.GetBytes(feature, "properties.wof:created")

	if !created.Exists() {

		feature, err = sjson.SetBytes(feature, "properties.wof:created", now)

		if err != nil {
			return nil, err
		}
	}

	return feature, nil
}
