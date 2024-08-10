package feature

import (
	"fmt"
)

type PropertyNotFoundErr struct {
	property string
}

func (e *PropertyNotFoundErr) Error() string {
	return fmt.Sprintf("'%s' property not found", e.property)
}

func (e *PropertyNotFoundErr) Property() string {
	return e.property
}

func PropertyNotFoundError(prop string) *PropertyNotFoundErr {
	return &PropertyNotFoundErr{
		property: prop,
	}
}

func IsPropertyNotFoundError(target error) bool {
	switch target.(type) {
	case *PropertyNotFoundErr:
		return true
	default:
		return false
	}
}
