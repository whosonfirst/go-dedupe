package properties

import (
	"context"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	id "github.com/whosonfirst/go-whosonfirst-id"
)

func EnsureWOFId(feature []byte, provider id.Provider) ([]byte, error) {

	// Eventually `ctx` should be part of the method signature but
	// should happen go-whosonfirst-export wide and will be a backwards
	// incompatible change

	ctx := context.Background()

	var err error

	var wof_id int64

	rsp := gjson.GetBytes(feature, "properties.wof:id")

	if rsp.Exists() {

		wof_id = rsp.Int()

	} else {

		i, err := provider.NewID(ctx)

		if err != nil {
			return nil, err
		}

		wof_id = i

		feature, err = sjson.SetBytes(feature, "properties.wof:id", wof_id)

		if err != nil {
			return nil, err
		}
	}

	id := gjson.GetBytes(feature, "id")

	if !id.Exists() {

		feature, err = sjson.SetBytes(feature, "id", wof_id)

		if err != nil {
			return nil, err
		}

	}

	return feature, nil
}
