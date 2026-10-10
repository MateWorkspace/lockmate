package domainmodelserror

import (
	"errors"
	"fmt"
)

type Type error

type Error struct {
	Message string
	Type    Type
	Source  error
}

func NewError(message string, errType Type, source error) error {
	return &Error{
		Message: message,
		Type:    errType,
		Source:  source,
	}
}

func (e *Error) Error() string {
	if e.Source != nil {
		return fmt.Sprintf("[%v] %s: %v", e.Type, e.Message, e.Source)
	}
	return fmt.Sprintf("[%v] %s", e.Type, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Source
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.Type, target)
}
