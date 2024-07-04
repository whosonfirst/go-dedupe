package dedupe

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/paulmach/orb/geojson"
)

func init() {

	// https://pkg.go.dev/github.com/paulmach/orb/geojson#pkg-variables
	// https://github.com/json-iterator/go
	//
	// "Even the most widely used json-iterator will severely degrade in generic (no-schema) or big-volume JSON serialization and deserialization."
	// https://github.com/bytedance/sonic/blob/main/INTRODUCTION.md
	//
	// I have not verified that claim either way but since we're not trafficing in "big-volume" JSON files
	// I am just going to see how this (json-iterator) goes for now.

	var c = jsoniter.Config{
		EscapeHTML:              true,
		SortMapKeys:             false,
		MarshalFloatWith6Digits: true,
	}.Froze()

	geojson.CustomJSONMarshaler = c
	geojson.CustomJSONUnmarshaler = c
}
