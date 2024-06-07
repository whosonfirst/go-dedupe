package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

func Repo(body []byte) (string, error) {

	rsp := gjson.GetBytes(body, "properties.wof:repo")

	if !rsp.Exists() {
		return "", fmt.Errorf("Missing wof:repo property")
	}

	repo := rsp.String()
	return repo, nil
}
