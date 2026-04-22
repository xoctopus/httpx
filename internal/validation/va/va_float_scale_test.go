package va_test

import (
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func TestFloatScaleVa(t *testing.T) {
	v, err := va.NewFloatScaleVa(must.NoErrorV(rule.Compile([]byte(`@float`))))
	Expect(t, err, Succeed())
	Expect(t, v.String(), Equal(""))
	Expect(t, v.Validate("10000000000.000000001"), Succeed())

	v, err = va.NewFloatScaleVa(must.NoErrorV(rule.Compile([]byte(`@float<1,2,3>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_FLOAT_SCALE))

	v, err = va.NewFloatScaleVa(must.NoErrorV(rule.Compile([]byte(`@float<@rule,2>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_FLOAT_SCALE))

	v, err = va.NewFloatScaleVa(must.NoErrorV(rule.Compile([]byte(`@float<abc,2>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_FLOAT_SCALE))

	v, err = va.NewFloatScaleVa(must.NoErrorV(rule.Compile([]byte(`@float<5,@rule>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_FLOAT_SCALE))

	v, err = va.NewFloatScaleVa(must.NoErrorV(rule.Compile([]byte(`@float<5,abc>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_FLOAT_SCALE))

	v, err = va.NewFloatScaleVa(must.NoErrorV(rule.Compile([]byte(`@float<5,8>`))))
	Expect(t, err, IsCodeError(va.ERROR__INVALID_FLOAT_SCALE))

	v, err = va.NewFloatScaleVa(must.NoErrorV(rule.Compile([]byte(`@float<5,2>`))))
	Expect(t, err, Succeed())
	Expect(t, v.String(), Equal("<5,2>"))

	Expect(t, v.Validate("1234.56"), IsCodeError(va.ERROR__OUT_OF_FLOAT_SCALE))
	Expect(t, v.Validate("1234"), Succeed())
	Expect(t, v.Validate("0"), Succeed())
	Expect(t, v.Validate("0.123"), IsCodeError(va.ERROR__OUT_OF_FLOAT_SCALE))
	Expect(t, v.Validate("3402823466385288598117041834845169254400.000000"), IsCodeError(va.ERROR__OUT_OF_FLOAT_SCALE))

	b := rule.NewBuilder("testing")
	r := must.NoErrorV(rule.Compile([]byte(`@float64<22,4>`)))
	v = must.NoErrorV(va.NewFloatScaleVa(r))
	v.BuildTo(b)
	Expect(t, b.Bytes(), Equal([]byte("@testing<22,4>")))
}
