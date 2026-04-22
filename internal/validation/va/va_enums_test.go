package va_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func TestEnumVa(t *testing.T) {
	r, err := rule.Compile([]byte("@int{1,2,3}"))
	Expect(t, err, Succeed())
	v, err := va.NewEnumVa[int](r)
	Expect(t, err, Succeed())
	Expect(t, v.Validate(10), IsCodeError(va.ERROR__OUT_OF_ENUMERATED_VALUES))
	Expect(t, v.Validate(2), Succeed())

	b := rule.NewBuilder("any")
	v.BuiltTo(b)
	Expect(t, b.Bytes(), Equal([]byte("@any{1,2,3}")))

	v = nil
	Expect(t, v.Validate(2), Succeed())

	r, err = rule.Compile([]byte("@int"))
	Expect(t, err, Succeed())
	v, err = va.NewEnumVa[int](r)
	Expect(t, err, Succeed())
	Expect(t, v, BeNil[*va.EnumVa[int]]())

	r, err = rule.Compile([]byte("@int{}"))
	Expect(t, err, Succeed())
	v, err = va.NewEnumVa[int](r)
	Expect(t, err, Succeed())
	Expect(t, v, BeNil[*va.EnumVa[int]]())

	r, err = rule.Compile([]byte("@int{%2}"))
	Expect(t, err, Succeed())
	v, err = va.NewEnumVa[int](r)
	Expect(t, err, Succeed())
	Expect(t, v, BeNil[*va.EnumVa[int]]())

	r, err = rule.Compile([]byte("@int{100.001}"))
	Expect(t, err, Succeed())
	v, err = va.NewEnumVa[int](r)
	Expect(t, err, IsCodeError(va.ERROR__INVALID_ENUM))
}
