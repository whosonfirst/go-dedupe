package properties

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	_ "log"

	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func EnsureSrcGeom(feature []byte) ([]byte, error) {

	path := "properties.src:geom"

	rsp := gjson.GetBytes(feature, path)

	if rsp.Exists() {
		return feature, nil
	}

	return sjson.SetBytes(feature, path, "unknown")
}

func EnsureGeomHash(feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, "geometry")

	if !rsp.Exists() {
		return nil, errors.New("missing geometry")
	}

	enc, err := json.Marshal(rsp.Value())

	if err != nil {
		return nil, err
	}

	hash := md5.Sum(enc)
	geom_hash := hex.EncodeToString(hash[:])

	return sjson.SetBytes(feature, "properties.wof:geomhash", geom_hash)
}

func EnsureGeomCoords(feature []byte) ([]byte, error) {

	// https://github.com/paulmach/orb/blob/master/geojson/feature.go
	// https://github.com/paulmach/orb/blob/master/planar/area.go

	var err error

	f, err := geojson.UnmarshalFeature(feature)

	if err != nil {
		return nil, err
	}

	centroid, area := planar.CentroidArea(f.Geometry)

	feature, err = sjson.SetBytes(feature, "properties.geom:latitude", centroid.Y())

	if err != nil {
		return nil, err
	}

	feature, err = sjson.SetBytes(feature, "properties.geom:longitude", centroid.X())

	if err != nil {
		return nil, err
	}

	feature, err = sjson.SetBytes(feature, "properties.geom:area", area)

	if err != nil {
		return nil, err
	}

	bounds := f.Geometry.Bound()

	min := bounds.Min
	max := bounds.Max

	minx := min.X()
	miny := min.Y()
	maxx := max.X()
	maxy := max.Y()

	bbox := []float64{
		minx,
		miny,
		maxx,
		maxy,
	}

	str_bbox := fmt.Sprintf("%.06f,%.06f,%.06f,%.06f", minx, miny, maxx, maxy)

	feature, err = sjson.SetBytes(feature, "properties.geom:bbox", str_bbox)

	if err != nil {
		return nil, err
	}

	feature, err = sjson.SetBytes(feature, "bbox", bbox)

	if err != nil {
		return nil, err
	}

	return feature, nil
}
