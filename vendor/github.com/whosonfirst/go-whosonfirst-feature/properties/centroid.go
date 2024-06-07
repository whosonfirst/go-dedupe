package properties

import (
	"fmt"
	"github.com/paulmach/orb"
	"github.com/tidwall/gjson"
)

func Centroid(body []byte) (*orb.Point, string, error) {

	props := []string{
		"lbl",
		"reversegeo",
		"mps",
		"geom",
	}

	var pt *orb.Point
	var source string

	for _, prefix := range props {

		path_lat := fmt.Sprintf("properties.%s:latitude", prefix)
		path_lon := fmt.Sprintf("properties.%s:longitude", prefix)

		rsp_lat := gjson.GetBytes(body, path_lat)
		rsp_lon := gjson.GetBytes(body, path_lon)

		if !rsp_lat.Exists() {
			continue
		}

		if !rsp_lon.Exists() {
			continue
		}

		pt = &orb.Point{rsp_lon.Float(), rsp_lat.Float()}
		source = prefix
		break
	}

	if pt == nil {
		pt = &orb.Point{0.0, 0.0}
		source = "nullisland"
	}

	return pt, source, nil
}
