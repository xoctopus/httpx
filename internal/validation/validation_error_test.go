package validation_test

import (
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"testing"

	"github.com/xoctopus/x/codex"
	. "github.com/xoctopus/x/testx"

	. "github.com/xoctopus/httpx/internal/validation"
)

func TestValidationError(t *testing.T) {
	Expect(t, IsValidationError(errors.New("")), BeFalse())
	Expect(t, IsValidationError(codex.New(ERROR__INPUT_VALUE)), BeTrue())
	Expect(t, IsValidationError(&PointerError{}), BeTrue())
	Expect(t, IsValidationError(nil), BeFalse())
}

func TestPointer(t *testing.T) {
	pointer := func(err error) jsontext.Pointer {
		x, _ := errors.AsType[*PointerError](err)
		return x.Pointer()
	}

	t.Run("WrapKeepsInnermost", func(t *testing.T) {
		err := WrapPointer(WrapPointer(errors.New("x"), "/items/0/age"), "/items")
		Expect(t, pointer(err), Equal[jsontext.Pointer]("/items/0/age"))
	})
	t.Run("PrefixNestedDocument", func(t *testing.T) {
		err := PrefixPointer(WrapPointer(errors.New("x"), "/items/0/age"), "/body")
		Expect(t, pointer(err), Equal[jsontext.Pointer]("/body/items/0/age"))
		Expect(t, err.Error(), Equal("x: [pointer:/body/items/0/age]"))
	})
	t.Run("PrefixRootPointer", func(t *testing.T) {
		err := PrefixPointer(WrapPointer(errors.New("x"), ""), "/body")
		Expect(t, pointer(err), Equal[jsontext.Pointer]("/body"))
	})
	t.Run("PrefixWithoutPointer", func(t *testing.T) {
		err := PrefixPointer(errors.New("x"), jsontext.Pointer("/query").AppendToken("ids"))
		Expect(t, pointer(err), Equal[jsontext.Pointer]("/query/ids"))
	})
	t.Run("PrefixThroughWrappers", func(t *testing.T) {
		inner := WrapPointer(errors.New("x"), "/a")
		err := PrefixPointer(fmt.Errorf("wrap: %w", inner), "/body")
		Expect(t, pointer(err), Equal[jsontext.Pointer]("/body/a"))
	})
	t.Run("PrefixJoinedBranches", func(t *testing.T) {
		a, b := WrapPointer(errors.New("a"), "/a"), WrapPointer(errors.New("b"), "/b")
		_ = PrefixPointer(errors.Join(a, b), "/body")
		Expect(t, pointer(a), Equal[jsontext.Pointer]("/body/a"))
		Expect(t, pointer(b), Equal[jsontext.Pointer]("/body/b"))
	})
	t.Run("Nil", func(t *testing.T) {
		Expect(t, WrapPointer(nil, "/a"), BeNil[error]())
		Expect(t, PrefixPointer(nil, "/a"), BeNil[error]())
	})
}
