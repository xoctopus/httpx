package va_test

import (
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func TestMultipleVa(t *testing.T) {
	v, err := va.NewMultipleVa[int](must.NoErrorV(rule.Compile([]byte("@int{1,2,3}"))))
	Expect(t, err, Succeed())
	Expect(t, v, BeNil[*va.MultipleVa[int]]())

	v, err = va.NewMultipleVa[int](must.NoErrorV(rule.Compile([]byte("@int"))))
	Expect(t, err, Succeed())
	Expect(t, v, BeNil[*va.MultipleVa[int]]())

	v, err = va.NewMultipleVa[int](must.NoErrorV(rule.Compile([]byte("@int{}"))))
	Expect(t, err, Succeed())
	Expect(t, v, BeNil[*va.MultipleVa[int]]())

	v, err = va.NewMultipleVa[int](must.NoErrorV(rule.Compile([]byte("@int{1}"))))
	Expect(t, err, Succeed())
	Expect(t, v, BeNil[*va.MultipleVa[int]]())

	v, err = va.NewMultipleVa[int](must.NoErrorV(rule.Compile([]byte("@int{%0}"))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_MULTIPLE))

	v, err = va.NewMultipleVa[int](must.NoErrorV(rule.Compile([]byte("@int{%0.001}"))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_MULTIPLE))

	v, err = va.NewMultipleVa[int](must.NoErrorV(rule.Compile([]byte("@int{%100}"))))
	Expect(t, err, Succeed())

	Expect(t, v.Validate(0, 0), Succeed())
	Expect(t, v.Validate(1000, 0), Succeed())
	Expect(t, v.Validate(1001, 0), IsCodeError(va.ERROR__NOT_MATCH_MULTIPLE))

	v = nil
	Expect(t, v.Validate(0, 0), Succeed())
	Expect(t, v.Validate(1000, 0), Succeed())
	Expect(t, v.Validate(1001, 0), Succeed())

	vUint, err := va.NewMultipleVa[uint](must.NoErrorV(rule.Compile([]byte("@uint{%2}"))))
	Expect(t, err, Succeed())
	Expect(t, vUint.Validate(0, 0), Succeed())
	Expect(t, vUint.Validate(4, 0), Succeed())
	Expect(t, vUint.Validate(7, 0), IsCodeError(va.ERROR__NOT_MATCH_MULTIPLE))

	vFloat, err := va.NewMultipleVa[float32](must.NoErrorV(rule.Compile([]byte("@float{%2.0002}"))))
	Expect(t, err, Succeed())
	Expect(t, vFloat.Validate(2.0002, 4), Succeed())
	Expect(t, vFloat.Validate(4.00041, 4), Succeed())
	Expect(t, vFloat.Validate(8.00082, 5), IsCodeError(va.ERROR__NOT_MATCH_MULTIPLE))
	Expect(t, vFloat.Validate(8.000802, 5), Succeed())

	b := rule.NewBuilder("any")
	vFloat.BuiltTo(b)
	Expect(t, b.Bytes(), Equal([]byte(`@any{%2.0002}`)))
}
