package properties

import (
	"github.com/sfomuseum/go-edtf"
	"github.com/tidwall/gjson"
)

func Inception(body []byte) string {

	rsp := gjson.GetBytes(body, "properties.edtf:inception")

	if !rsp.Exists() {
		return edtf.UNKNOWN
	}

	return rsp.String()
}

func Cessation(body []byte) string {

	rsp := gjson.GetBytes(body, "properties.edtf:cessation")

	if !rsp.Exists() {
		return edtf.UNKNOWN
	}

	return rsp.String()
}

func Deprecated(body []byte) string {

	rsp := gjson.GetBytes(body, "properties.edtf:deprecated")
	return rsp.String()
}
