package properties

import (
	"github.com/tidwall/gjson"
	"strings"
)

func Names(body []byte) map[string][]string {

	names_map := make(map[string][]string)

	r := gjson.GetBytes(body, "properties")

	if !r.Exists() {
		return names_map
	}

	for k, v := range r.Map() {

		if !strings.HasPrefix(k, "name:") {
			continue
		}

		if !v.Exists() {
			continue
		}

		name := strings.Replace(k, "name:", "", 1)
		names := make([]string, 0)

		for _, n := range v.Array() {
			names = append(names, n.String())
		}

		names_map[name] = names
	}

	return names_map
}
