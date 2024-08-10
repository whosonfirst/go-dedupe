package format

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/tidwall/pretty"
)

// Feature represents a WOF Feature, ready to be encoded to JSON
type Feature struct {
	Type       string      `json:"type"`
	ID         int64       `json:"id"`
	Properties interface{} `json:"properties"`
	Bbox       interface{} `json:"bbox,omitempty"`
	Geometry   interface{} `json:"geometry"`
}

// two space indent
const indent = "  "

// FormatFeature transforms a byte array `b` into a correctly formatted WOF file
func FormatBytes(b []byte) ([]byte, error) {
	var f *Feature
	err := json.Unmarshal(b, &f)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal bytes in to Feature, %w", err)
	}

	return FormatFeature(f)
}

// FormatFeature transforms a Feature into a correctly formatted WOF file
func FormatFeature(feature *Feature) ([]byte, error) {
	var buf bytes.Buffer

	_, err := buf.WriteString("{\n")
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "id", feature.ID, true, false)
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "type", feature.Type, true, false)
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "properties", feature.Properties, true, false)
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "bbox", feature.Bbox, true, false)
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "geometry", feature.Geometry, false, true)
	if err != nil {
		return buf.Bytes(), err
	}

	_, err = buf.WriteString("\n}\n")
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func writeKey(buf *bytes.Buffer, key string, value interface{}, usePretty, lastLine bool) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if usePretty {
		prefix := indent
		prettyOpts := &pretty.Options{Indent: indent, SortKeys: true, Prefix: prefix}
		valueJSON = pretty.PrettyOptions(valueJSON, prettyOpts)
		// Trim the newline that comes back from pretty, so we can control it last
		valueJSON = valueJSON[:len(valueJSON)-1]
		// Trim the first prefix
		valueJSON = valueJSON[len(indent):]
	} else {
		valueJSON = pretty.Ugly(valueJSON)
	}

	trailing := ",\n"
	if lastLine {
		trailing = ""
	}

	_, err = buf.WriteString(indent)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(buf, "\"%s\": %s%s", key, valueJSON, trailing)
	return err
}
