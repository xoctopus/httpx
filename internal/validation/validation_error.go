package validation

import (
	"errors"
	"fmt"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	if codex.Is[rule.Error](err) || codex.Is[va.Error](err) || codex.Is[Error](err) {
		return true
	}
	_, hasLocationError := errors.AsType[*PositionError](err)
	_, hasPositionError := errors.AsType[*LocationError](err)
	return hasLocationError || hasPositionError
}

func WrapPositionError[Pos ~string](err error, position Pos) error {
	if len(position) == 0 {
		position = "unknown-position"
	}
	if err == nil {
		return nil
	}
	return &PositionError{err, string(position)}
}

type PositionError struct {
	err error
	pos string
}

func (e *PositionError) Error() string {
	if x, ok := errors.AsType[*LocationError](e.err); ok {
		return fmt.Sprintf("%s: [loc:%s pos:%s]", e.Unwrap(), x.Location(), e.pos)
	}
	return fmt.Sprintf("%s: [pos:%s]", e.err, e.pos)
}

func (e *PositionError) Unwrap() error {
	if x, ok := errors.AsType[*LocationError](e.err); ok {
		return x.Unwrap()
	}
	return e.err
}

func (e *PositionError) Position() string {
	return e.pos
}

func WrapLocationError(err error, location string) error {
	if len(location) == 0 {
		location = "unknown-location"
	}
	if err == nil {
		return nil
	}
	return &LocationError{err, location}
}

type LocationError struct {
	err error
	loc string
}

func (e *LocationError) Error() string {
	if x, ok := errors.AsType[*PositionError](e.err); ok {
		return fmt.Sprintf("%s: [loc:%s pos:%s]", e.Unwrap(), e.loc, x.Position())
	}
	return fmt.Sprintf("%s: [loc:%s]", e.err, e.loc)
}

func (e *LocationError) Unwrap() error {
	if x, ok := errors.AsType[*PositionError](e.err); ok {
		return x.Unwrap()
	}
	return e.err
}

func (e *LocationError) Location() string {
	return e.loc
}
