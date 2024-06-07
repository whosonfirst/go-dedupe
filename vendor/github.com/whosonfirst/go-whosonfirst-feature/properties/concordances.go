package properties

import (
	"github.com/tidwall/gjson"
)

func Concordances(body []byte) map[string]interface{} {

	concordances := make(map[string]interface{})

	rsp := gjson.GetBytes(body, "properties.wof:concordances")

	for k, v := range rsp.Map() {
		concordances[k] = v.Value()
	}

	return concordances
}
