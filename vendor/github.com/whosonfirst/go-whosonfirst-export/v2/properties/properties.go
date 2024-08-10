package properties

import (
	"fmt"
)

func EnsureRequired(feature []byte) ([]byte, error) {

	var err error

	feature, err = EnsureName(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure wof:name, %w", err)
	}

	feature, err = EnsurePlacetype(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure placetype, %w", err)
	}

	feature, err = EnsureGeom(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure geometry, %w", err)
	}

	return feature, nil
}

func EnsureGeom(feature []byte) ([]byte, error) {

	var err error

	feature, err = EnsureSrcGeom(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure src:geom, %w", err)
	}

	feature, err = EnsureGeomHash(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure geom:hash, %w", err)
	}

	feature, err = EnsureGeomCoords(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure geometry coordinates, %w", err)
	}

	return feature, nil
}

func EnsureTimestamps(feature []byte) ([]byte, error) {

	var err error

	feature, err = EnsureCreated(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure wof:created, %w", err)
	}

	feature, err = EnsureLastModified(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure wof:lastmodified, %w", err)
	}

	return feature, nil
}
