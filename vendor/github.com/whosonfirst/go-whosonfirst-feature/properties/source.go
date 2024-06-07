package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

func Source(body []byte) (string, error) {

	var source string

	possible := []string{
		"properties.src:alt_label",
		"properties.src:geom",
	}

	for _, path := range possible {

		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() {
			source = rsp.String()
			break
		}
	}

	if source == "" {
		return "", fmt.Errorf("Missing src:geom or src:alt_label property")
	}

	return source, nil
}
