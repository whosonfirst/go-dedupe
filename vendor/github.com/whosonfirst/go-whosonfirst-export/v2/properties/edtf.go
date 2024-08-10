package properties

import (
	"github.com/sfomuseum/go-edtf"
	"github.com/sfomuseum/go-edtf/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const dateFmt string = "2006-01-02"

func EnsureEDTF(feature []byte) ([]byte, error) {
	var err error

	feature, err = EnsureInception(feature)
	if err != nil {
		return nil, err
	}

	feature, err = EnsureCessation(feature)
	if err != nil {
		return nil, err
	}

	return feature, nil
}

func EnsureInception(feature []byte) ([]byte, error) {
	path := "properties.edtf:inception"
	lowerPath := "properties.date:inception_lower"
	upperPath := "properties.date:inception_upper"

	return updatePath(feature, path, upperPath, lowerPath)
}

func EnsureCessation(feature []byte) ([]byte, error) {
	path := "properties.edtf:cessation"
	lowerPath := "properties.date:cessation_lower"
	upperPath := "properties.date:cessation_upper"

	return updatePath(feature, path, upperPath, lowerPath)
}

func updatePath(feature []byte, path string, upperPath string, lowerPath string) ([]byte, error) {
	property := gjson.GetBytes(feature, path)

	if !property.Exists() {
		return setProperties(feature, edtf.UNKNOWN, path, upperPath, lowerPath)
	}

	edtfStr := property.String()

	return setProperties(feature, edtfStr, path, upperPath, lowerPath)
}

func setProperties(feature []byte, edtfStr string, path string, upperPath, lowerPath string) ([]byte, error) {
	feature, err := sjson.SetBytes(feature, path, edtfStr)
	if err != nil {
		return nil, err
	}

	switch edtfStr {
	case edtf.UNKNOWN, edtf.OPEN:
		return removeUpperLower(feature, upperPath, lowerPath)
	case edtf.UNKNOWN_2012:
		return setProperties(feature, edtf.UNKNOWN, path, upperPath, lowerPath)
	default:
		return setUpperLower(feature, edtfStr, upperPath, lowerPath)
	}
}

func setUpperLower(feature []byte, edtfStr string, upperPath string, lowerPath string) ([]byte, error) {
	dt, err := parser.ParseString(edtfStr)
	if err != nil {
		return nil, err
	}

	lowerTime, err := dt.Lower()
	if err != nil {
		return nil, err
	}

	feature, err = sjson.SetBytes(feature, lowerPath, lowerTime.Format(dateFmt))
	if err != nil {
		return nil, err
	}

	upperTime, err := dt.Upper()
	if err != nil {
		return nil, err
	}

	feature, err = sjson.SetBytes(feature, upperPath, upperTime.Format(dateFmt))
	if err != nil {
		return nil, err
	}

	return feature, nil
}

func removeUpperLower(feature []byte, upperPath string, lowerPath string) ([]byte, error) {
	feature, err := sjson.DeleteBytes(feature, upperPath)
	if err != nil {
		return nil, err
	}

	feature, err = sjson.DeleteBytes(feature, lowerPath)
	if err != nil {
		return nil, err
	}

	return feature, nil
}
