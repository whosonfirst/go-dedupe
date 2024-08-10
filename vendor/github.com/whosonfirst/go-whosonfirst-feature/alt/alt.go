// package alt provides methods for working with "alternate" geometry records
package alt

import (
	"github.com/tidwall/gjson"
)

func IsAlt(body []byte) bool {

	allowed_properties := []string{
		// this is the new new but won't "work" until we backfill all
		// 26M files and the export tools to set this property
		// (20190821/thisisaaronland)
		"properties.src:alt_label",
		// SFO syntax (initial proposal)
		"properties.wof:alt_label",
	}

	for _, path := range allowed_properties {

		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() && rsp.String() != "" {
			return true
		}
	}

	// we used to test that wof:parent_id wasn't -1 but that's a bad test since
	// plenty of stuff might have a parent ID of -1 and really what we want to
	// test is the presence of the property not the value
	// (20190821/thisisaaronland)

	return false
}
