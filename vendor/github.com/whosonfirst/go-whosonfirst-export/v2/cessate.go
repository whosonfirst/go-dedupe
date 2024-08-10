package export

import (
	"context"
	"fmt"
	"time"
)

// CessateRecord will assign the relevant properties to make 'old_body' as cessated (ceased) using the current time.
func CessateRecord(ctx context.Context, ex Exporter, old_body []byte) ([]byte, error) {
	t := time.Now()
	return CessateRecordWithTime(ctx, ex, t, old_body)
}

// CessateRecordWithTime will assign the relevant properties to make 'old_body' as cessated (ceased) using the time defined by 't'.
func CessateRecordWithTime(ctx context.Context, ex Exporter, t time.Time, old_body []byte) ([]byte, error) {

	to_update := map[string]interface{}{
		"properties.edtf:cessation": t.Format("2006-01-02"),
		"properties.mz:is_current":  0,
	}

	new_body, err := AssignProperties(ctx, old_body, to_update)

	if err != nil {
		return nil, fmt.Errorf("Failed to assign properties, %w", err)
	}

	new_body, err = ex.Export(ctx, new_body)

	if err != nil {
		return nil, fmt.Errorf("Failed to export body, %w", err)
	}

	return new_body, nil
}
