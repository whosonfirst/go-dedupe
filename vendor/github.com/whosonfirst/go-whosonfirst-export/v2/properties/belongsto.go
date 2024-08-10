package properties

import (
	"errors"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func EnsureBelongsTo(feature []byte) ([]byte, error) {

	belongsto := make([]int64, 0)

	wofid_rsp := gjson.GetBytes(feature, "properties.wof:id")

	if !wofid_rsp.Exists() {
		return nil, errors.New("missing properties.wof:id")
	}

	wofid := wofid_rsp.Int()

	// Load the existing belongsto array, if it exists

	belongsToRsp := gjson.GetBytes(feature, "properties.wof:belongsto")

	if belongsToRsp.Exists() {

		belongsToRsp.ForEach(func(key gjson.Result, value gjson.Result) bool {
			if value.Type == gjson.Number {
				id := value.Int()
				belongsto = append(belongsto, id)
			}

			return true
		})
	}

	rsp := gjson.GetBytes(feature, "properties.wof:hierarchy")

	if rsp.Exists() {

		ids := make([]int64, 0)

		for _, h := range rsp.Array() {
			h.ForEach(func(key gjson.Result, value gjson.Result) bool {

				if value.Type == gjson.Number {

					id := value.Int()

					if id > 0 && id != wofid {
						ids = append(ids, id)
					}
				}

				return true
			})
		}

		// Add all the IDs we've not seen before
		for _, id := range ids {
			if !sliceContains(belongsto, id) {
				belongsto = append(belongsto, id)
			}
		}

		// Remove all the IDs we no longer want in the list - in reverse,
		// because Golang.
		for i := len(belongsto) - 1; i >= 0; i-- {
			id := belongsto[i]

			if !sliceContains(ids, id) {
				belongsto = append(belongsto[:i], belongsto[i+1:]...)
			}
		}
	}

	tz_rsp := gjson.GetBytes(feature, "properties.wof:timezones")

	if tz_rsp.Exists() {

		for _, i := range tz_rsp.Array() {

			id := i.Int()
			if !sliceContains(belongsto, id) {
				belongsto = append(belongsto, id)
			}
		}
	}

	return sjson.SetBytes(feature, "properties.wof:belongsto", belongsto)
}

func sliceContains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
