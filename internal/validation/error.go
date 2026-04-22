package validation

import (
	"fmt"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

// Error validation error
// +genx:code
type Error int8

const (
	ERROR_UNDEFINED           Error = iota
	ERROR__STRING_LENGTH_MODE       // invalid string length mode, expect 'byte' or 'rune'
	ERROR__NOT_MATCH_REGEXP         // invalid string not match regexp
	ERROR__SLICE_PARAM              // invalid string parameter
	ERROR__MAP_PARAM                // invalid map parameter
	ERROR__INPUT_TYPE               // invalid input: invalid type
	ERROR__INPUT_VALUE              // invalid input: invalid value
	ERROR__MISSING_REQUIRED         // invalid input: missing required
)

func IsValidationError(err error) bool {
	_, ok1 := codex.Is[rule.Error](err)
	_, ok2 := codex.Is[va.Error](err)
	_, ok3 := codex.Is[Error](err)

	return ok1 || ok2 || ok3
}

func WrapPosition[P ~string](err error, position P) error {
	if err == nil || len(position) == 0 {
		return nil
	}
	return &HasPosition{err, string(position)}
}

func WrapLocation(err error, location string) error {
	if err == nil || len(location) == 0 {
		return nil
	}
	return &HasLocation{err, location}
}

func WrapLocationPosition(err error, location, position string) error {
	if err == nil || len(location) == 0 && len(position) == 0 {
		return nil
	}

	if len(location) > 0 && len(position) > 0 {
		return &HasLocationPosition{err, location, position}
	}

	if len(location) > 0 {
		return &HasLocation{err, location}
	}

	return &HasPosition{err, position}
}

type HasPosition struct {
	err error
	pos string
}

func (e *HasPosition) Error() string {
	return fmt.Sprintf("%s: at `%s`", e.err, e.pos)
}

func (e *HasPosition) Unwrap() error {
	return e.err
}

func (e *HasPosition) Position() string {
	return e.pos
}

type HasLocation struct {
	err error
	loc string
}

func (e *HasLocation) Error() string {
	return fmt.Sprintf("%s: in `%s`", e.err, e.loc)
}

func (e *HasLocation) Unwrap() error {
	return e.err
}

func (e *HasLocation) Location() string {
	return e.loc
}

type HasLocationPosition struct {
	err error
	loc string
	pos string
}

func (e *HasLocationPosition) Error() string {
	return fmt.Sprintf("%s: in `%s`", e.err, e.loc)
}

func (e *HasLocationPosition) Unwrap() error {
	return e.err
}

func (e *HasLocationPosition) Location() string {
	return e.loc
}

func (e *HasLocationPosition) Position() string {
	return e.pos
}
