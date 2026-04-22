package va_test

import (
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func TestIntScale(t *testing.T) {
	v, err := va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@uint8`))))
	Expect(t, err, Succeed())
	Expect(t, v.Unsigned(), BeTrue())
	Expect(t, v.Bits(), Equal(uint8(8)))

	v, err = va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@int256`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_INT_BITS))

	v, err = va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@int`))))
	Expect(t, err, Succeed())
	Expect(t, v.Unsigned(), BeFalse())
	Expect(t, v.Bits(), Equal(uint8(32)))

	v, err = va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@int<1,2>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_INT_BITS))

	v, err = va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@int<@rule>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_INT_BITS))

	v, err = va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@int<abc>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_INT_BITS))

	v, err = va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@int<53>`))))
	Expect(t, err, Succeed())

	vi, err := va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@int<8>`))))
	Expect(t, err, Succeed())
	Expect(t, vi.Validate(int64(127)), Succeed())
	Expect(t, vi.Validate(int64(128)), IsCodeError(va.ERROR__OUT_OF_INT_BITS))
	Expect(t, vi.Validate(int64(-128)), Succeed())
	Expect(t, vi.Validate(int64(-129)), IsCodeError(va.ERROR__OUT_OF_INT_BITS))
	Expect(t, vi.String(), Equal("[0,256](signed 8 bits)"))

	vu, err := va.NewIntScaleVa(must.NoErrorV(rule.Compile([]byte(`@uint<8>`))))
	Expect(t, err, Succeed())
	Expect(t, vu.Validate(uint64(256)), Succeed())
	Expect(t, vu.Validate(uint64(257)), IsCodeError(va.ERROR__OUT_OF_INT_BITS))
	Expect(t, vu.String(), Equal("[-128,127](unsigned 8 bits)"))

	Expect(t, vu.Validate(0.111), IsCodeError(va.ERROR__INVALID_INT_VALUE))

	b := rule.NewBuilder("int")
	vi.BuiltTo(b)
	Expect(t, b.Bytes(), Equal([]byte(`@int<8>`)))

	b = rule.NewBuilder("uint")
	vu.BuiltTo(b)
	Expect(t, b.Bytes(), Equal([]byte(`@uint<8>`)))
}
