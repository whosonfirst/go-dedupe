package properties

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
)

func Hierarchies(body []byte) []map[string]int64 {

	rsp := gjson.GetBytes(body, "properties.wof:hierarchy")

	if !rsp.Exists() {
		return nil
	}

	hierarchies := make([]map[string]int64, 0)

	for _, h := range rsp.Array() {

		dict := make(map[string]int64)

		for k, v := range h.Map() {
			dict[k] = v.Int()
		}

		hierarchies = append(hierarchies, dict)
	}

	return hierarchies
}

// MergeHierarchies returns the unique set of hierarchies defined in 'features'.
func MergeHierarchies(features ...[]byte) ([]map[string]int64, error) {

	tmp := make(map[string]map[string]int64)

	for i, body := range features {

		hierarchies := Hierarchies(body)

		if hierarchies == nil {
			continue
		}

		for j, h := range hierarchies {

			hash, err := hash_interface(h)

			if err != nil {
				return nil, fmt.Errorf("Failed to hash hierarchy at index %d for feature at index %d, %w", j, i, err)
			}

			tmp[hash] = h
		}
	}

	hierarchies := make([]map[string]int64, 0)

	for _, h := range tmp {
		hierarchies = append(hierarchies, h)
	}

	return hierarchies, nil
}

// hash_interface will return the SHA-256 hash for the JSON-encoding of 'i'.
func hash_interface(i interface{}) (string, error) {

	enc_i, err := json.Marshal(i)

	if err != nil {
		return "", fmt.Errorf("Failed to marshal interface, %w", err)
	}

	h := sha256.New()
	h.Write(enc_i)

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
