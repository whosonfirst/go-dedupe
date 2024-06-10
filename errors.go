package dedupe

import (
	"errors"
	"fmt"
)

type InvalidRecordError struct {
	id    string
	error error
}

func InvalidRecord(id string, error error) *InvalidRecordError {

	e := &InvalidRecordError{
		id:    id,
		error: error,
	}

	return e
}

func IsInvalidRecordError(e error) bool {
	var invalid *InvalidRecordError
	return errors.As(e, &invalid)
}

func (e *InvalidRecordError) Error() string {
	return fmt.Sprintf("%s is an invalid record, %v", e.id, e.error)
}

func (e *InvalidRecordError) String() string {
	return e.Error()
}
