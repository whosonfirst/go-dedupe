package export

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func EnsureProperties(ctx context.Context, body []byte, to_ensure map[string]interface{}) ([]byte, error) {

	to_assign := make(map[string]interface{})

	for path, v := range to_ensure {

		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() {
			continue
		}

		to_assign[path] = v
	}

	return AssignProperties(ctx, body, to_assign)
}

func AssignProperties(ctx context.Context, body []byte, to_assign map[string]interface{}) ([]byte, error) {

	var err error

	for path, v := range to_assign {

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return nil, err
		}
	}

	return body, nil
}

func AssignPropertiesIfChanged(ctx context.Context, body []byte, to_assign map[string]interface{}) (bool, []byte, error) {

	var err error

	changed := false

	for path, v := range to_assign {

		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() {

			old, err := json.Marshal(rsp.Value())

			if err != nil {
				return changed, nil, err
			}

			new, err := json.Marshal(v)

			if bytes.Equal(old, new) {
				continue
			}

			if err != nil {
				return changed, nil, err
			}
		}

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return changed, nil, err
		}

		changed = true
	}

	return changed, body, nil
}

func RemoveProperties(ctx context.Context, body []byte, to_remove []string) ([]byte, error) {

	var err error

	for _, path := range to_remove {

		body, err = sjson.DeleteBytes(body, path)

		if err != nil {
			return nil, err
		}
	}

	return body, nil
}
