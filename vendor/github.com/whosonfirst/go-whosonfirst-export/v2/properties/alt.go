package properties

import (
	"fmt"

	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func EnsureSourceAltLabel(feature []byte) ([]byte, error) {

	label, err := wof_properties.AltLabel(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive src:alt_label, %w", err)
	}

	if label == "" {
		return nil, fmt.Errorf("Invalid or empty src:alt_label property")
	}

	return feature, nil
}
