package validation

import (
	"encoding/json/jsontext"
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
	_, ok := errors.AsType[*PointerError](err)
	return ok
}

// WrapPointer attaches pointer p to err. p is a JSON Pointer (RFC 6901) in the
// same document as the pointers already in err, so an existing pointer is kept
// since it is the innermost and the most precise one.
func WrapPointer(err error, p jsontext.Pointer) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*PointerError](err); ok {
		return err
	}
	return &PointerError{err: err, ptr: p}
}

// PrefixPointer prepends prefix to the pointers in err. It is used when err
// comes from a nested document whose root is located by prefix in the outer
// document, eg: a request body is located by `/body` in the request.
func PrefixPointer(err error, prefix jsontext.Pointer) error {
	if err == nil || prefix == "" {
		return err
	}
	if prefixPointer(err, prefix) {
		return err
	}
	return &PointerError{err: err, ptr: prefix}
}

func prefixPointer(err error, prefix jsontext.Pointer) bool {
	switch x := err.(type) {
	case *PointerError:
		x.ptr = prefix + x.ptr
		return true
	case interface{ Unwrap() error }:
		if e := x.Unwrap(); e != nil {
			return prefixPointer(e, prefix)
		}
	case interface{ Unwrap() []error }:
		prefixed := false
		for _, e := range x.Unwrap() {
			if e != nil && prefixPointer(e, prefix) {
				prefixed = true
			}
		}
		return prefixed
	}
	return false
}

// PointerError locates err by a JSON Pointer (RFC 6901) in the request document.
type PointerError struct {
	err error
	ptr jsontext.Pointer
}

func (e *PointerError) Error() string {
	return fmt.Sprintf("%s: [pointer:%s]", e.err, e.ptr)
}

func (e *PointerError) Unwrap() error {
	return e.err
}

func (e *PointerError) Pointer() jsontext.Pointer {
	return e.ptr
}
