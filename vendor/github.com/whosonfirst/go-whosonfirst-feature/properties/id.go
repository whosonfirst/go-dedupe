package properties

import (
	"fmt"
	"regexp"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature"
	"github.com/whosonfirst/go-whosonfirst-feature/constants"
)

var re_wofid = regexp.MustCompile(`^\-?\d+$`)

func Id(body []byte) (int64, error) {

	rsp := gjson.GetBytes(body, "properties.wof:id")

	if !rsp.Exists() {
		return 0, feature.PropertyNotFoundError("wof:id")
	}

	if !re_wofid.MatchString(rsp.Raw) {
		return constants.UNKNOWN, fmt.Errorf("Invalid wof:id '%s'", rsp.Raw)
	}

	wof_id := rsp.Int()

	if wof_id < 0 {

		switch wof_id {
		case constants.MULTIPLE_PARENTS, constants.MULTIPLE_NEIGHBOURHOODS, constants.ITS_COMPLICATED, constants.UNKNOWN:
			// pass
		default:
			return constants.UNKNOWN, fmt.Errorf("Invalid or unrecognized ID value (%d)", wof_id)
		}
	}

	return wof_id, nil
}
