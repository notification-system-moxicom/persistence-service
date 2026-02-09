package errors

import (
	"errors"
	"fmt"
)

type ConfigurationError struct {
	msg string
	err error
}

func NewConfigurationError(msg string, err error) ConfigurationError {
	return ConfigurationError{msg: msg, err: err}
}

func (e ConfigurationError) Error() string {
	if e.err == nil {
		return e.msg
	}

	return fmt.Sprintf("%s: %s", e.msg, e.err.Error())
}

func (e ConfigurationError) Is(err error) bool {
	var configurationError ConfigurationError

	ok := errors.As(err, &configurationError)

	return ok
}

type InvalidMessageError struct {
	msg string
	err error
}

func NewInvalidMessageError(msg string, err error) InvalidMessageError {
	return InvalidMessageError{msg: msg, err: err}
}

func (e InvalidMessageError) Error() string {
	if e.err == nil {
		return e.msg
	}

	return fmt.Sprintf("%s: %s", e.msg, e.err.Error())
}

func (e InvalidMessageError) Is(err error) bool {
	_, ok := err.(InvalidMessageError)

	return ok
}

type ValidationError struct {
	msg     string
	details []string
}

func NewValidationError(msg string, details ...string) ValidationError {
	return ValidationError{msg: msg, details: details}
}

func (e ValidationError) Error() string {
	return e.msg
}

func (e ValidationError) Details() []string {
	return e.details
}

func (e ValidationError) Is(err error) bool {
	var validationError ValidationError

	ok := errors.As(err, &validationError)

	return ok
}
