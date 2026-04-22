package va_test

import (
	"math"
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func TestLengthVa(t *testing.T) {
	for _, c := range []struct {
		rule string
		text string
	}{
		{`@string[10]`, `[10]`},
		{`@array[,10]`, `[0,10]`},
		{`@map<@string,@int>(3,]`, `(3,]`},
	} {
		r, err := rule.Compile([]byte(c.rule))
		Expect(t, err, Succeed())

		v, err := va.NewLengthVa(r)
		Expect(t, err, Succeed())

		Expect(t, v.String(), Equal(c.text))
	}

	t.Run("Failed", func(t *testing.T) {
		for _, c := range []string{
			`@string[-10]`,
			`@array[,abc]`,
			`@slice[def,]`,
			`@map<@string,@int>(3,2]`,
		} {
			r, err := rule.Compile([]byte(c))
			Expect(t, err, Succeed())

			_, err = va.NewLengthVa(r)
			Expect(t, err, IsCodeError(va.ERROR__INVALID_LENGTH_RANGE))
		}
	})

	t.Run("Ignored", func(t *testing.T) {
		for _, c := range []string{
			`@array()`,
			`@string[]`,
			`@array[,]`,
			`@map<@string,@int>[]`,
		} {
			r, err := rule.Compile([]byte(c))
			Expect(t, err, Succeed())

			v, err := va.NewLengthVa(r)
			Expect(t, err, Succeed())
			Expect(t, v, BeNil[*va.LengthVa]())
		}
	})

	t.Run("String", func(t *testing.T) {
		for _, c := range []struct {
			rule string
			text string
		}{
			{rule: "@string", text: ""},
			{rule: "@string[10]", text: "[10]"},
			{rule: "@string[,11]", text: "[0,11]"},
			{rule: "@string[0,]", text: "[0,]"},
			{rule: "@string(,13)", text: "(0,13)"},
			{rule: "@string(0,)", text: "(0,)"},
		} {
			r, err := rule.Compile([]byte(c.rule))
			Expect(t, err, Succeed())
			v, err := va.NewLengthVa(r)
			Expect(t, err, Succeed())

			Expect(t, v.String(), Equal(c.text))
		}
	})

	t.Run("Validate", func(t *testing.T) {
		v, _ := va.NewLengthVa(must.NoErrorV(rule.Compile([]byte(`@any[10]`))))
		Expect(t, v.Validate(1), IsCodeError(va.ERROR__OUT_OF_LENGTH))
		Expect(t, v.Validate(10), Succeed())

		v, _ = va.NewLengthVa(must.NoErrorV(rule.Compile([]byte(`@any[10,]`))))
		Expect(t, v.Validate(9), IsCodeError(va.ERROR__OUT_OF_LENGTH))
		Expect(t, v.Validate(10), Succeed())

		// v = &va.LengthVa{min: uint(10), exMin: true}
		v, _ = va.NewLengthVa(must.NoErrorV(rule.Compile([]byte(`@any(10,]`))))
		Expect(t, v.Validate(9), IsCodeError(va.ERROR__OUT_OF_LENGTH))
		Expect(t, v.Validate(10), IsCodeError(va.ERROR__OUT_OF_LENGTH))
		Expect(t, v.Validate(11), Succeed())

		// v = &va.LengthVa{max: new(uint(10))}
		v, _ = va.NewLengthVa(must.NoErrorV(rule.Compile([]byte(`@any[,10]`))))
		Expect(t, v.Validate(9), Succeed())
		Expect(t, v.Validate(10), Succeed())
		Expect(t, v.Validate(11), IsCodeError(va.ERROR__OUT_OF_LENGTH))

		// v = &va.LengthVa{max: new(uint(10)), exMax: true}
		v, _ = va.NewLengthVa(must.NoErrorV(rule.Compile([]byte(`@any[,10)`))))
		Expect(t, v.Validate(9), Succeed())
		Expect(t, v.Validate(10), IsCodeError(va.ERROR__OUT_OF_LENGTH))
		Expect(t, v.Validate(11), IsCodeError(va.ERROR__OUT_OF_LENGTH))

		v = nil
		Expect(t, v.Validate(10), Succeed())
		Expect(t, v.Validate(math.MaxUint), Succeed())
		Expect(t, v.Validate(0), Succeed())

		b := rule.NewBuilder("any")
		b.SetLengthMode(true)
		v, _ = va.NewLengthVa(b)
		Expect(t, v, BeNil[*va.LengthVa]())
	})

	t.Run("Builder", func(t *testing.T) {
		b := rule.NewBuilder("testing")
		r := must.NoErrorV(rule.Compile([]byte(`@testing[10]`)))
		v := must.NoErrorV(va.NewLengthVa(r))
		v.BuiltTo(b)
		Expect(t, b.Bytes(), Equal([]byte(`@testing[10]`)))

		b = rule.NewBuilder("testing")
		v = nil
		v.BuiltTo(b)
		Expect(t, b.Bytes(), Equal([]byte(`@testing`)))

		b = rule.NewBuilder("testing")
		r = must.NoErrorV(rule.Compile([]byte(`@testing[100,200]`)))
		v = must.NoErrorV(va.NewLengthVa(r))
		v.BuiltTo(b)
		Expect(t, r.Bytes(), Equal([]byte(`@testing[100,200]`)))
	})
}
