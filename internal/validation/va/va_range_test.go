package va_test

import (
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func TestRangeVa(t *testing.T) {
	for _, c := range []struct {
		rule string
		text string
	}{
		{`@int[10,11]`, `[10,11]`},
		{`@int[,10]`, `[,10]`},
		{`@int(3,]`, `(3,]`},
	} {
		r, err := rule.Compile([]byte(c.rule))
		Expect(t, err, Succeed())

		v, err := va.NewRangeVa[int](r)
		Expect(t, err, Succeed())

		Expect(t, v.String(), Equal(c.text))
	}

	t.Run("Failed", func(t *testing.T) {
		for _, c := range []string{
			`@int[-10]`,
			`@int[,abc]`,
			`@int[100.001,]`,
			`@int(3,2]`,
		} {
			r, err := rule.Compile([]byte(c))
			Expect(t, err, Succeed())

			_, err = va.NewRangeVa[int](r)
			Expect(t, err, IsCodeError(va.ERROR__INVALID_VALUE_RANGE))
		}
	})

	t.Run("Ignored", func(t *testing.T) {
		for _, c := range []string{
			`@int()`,
			`@int[]`,
			`@int[,]`,
		} {
			r, err := rule.Compile([]byte(c))
			Expect(t, err, Succeed())

			v, err := va.NewRangeVa[int](r)
			Expect(t, err, Succeed())
			Expect(t, v, BeNil[*va.RangeVa[int]]())
		}
	})

	t.Run("String", func(t *testing.T) {
		for _, c := range []struct {
			rule string
			text string
		}{
			{rule: "@int", text: ""},
			{rule: "@int[10,15]", text: "[10,15]"},
			{rule: "@int[,11]", text: "[,11]"},
			{rule: "@int[0,]", text: "[0,]"},
			{rule: "@int(,13)", text: "(,13)"},
			{rule: "@int(0,)", text: "(0,)"},
		} {
			r, err := rule.Compile([]byte(c.rule))
			Expect(t, err, Succeed())
			v, err := va.NewRangeVa[int](r)
			Expect(t, err, Succeed())

			Expect(t, v.String(), Equal(c.text))
		}
	})

	t.Run("Validate", func(t *testing.T) {
		r := must.NoErrorV(rule.Compile([]byte(`@int[10,100)`)))
		v := must.NoErrorV(va.NewRangeVa[int](r))
		Expect(t, v.Validate(9), IsCodeError(va.ERROR__OUT_OF_VALUE_RANGE))
		Expect(t, v.Validate(10), Succeed())
		Expect(t, v.Validate(99), Succeed())
		Expect(t, v.Validate(100), IsCodeError(va.ERROR__OUT_OF_VALUE_RANGE))

		v = nil
		Expect(t, v.Validate(9), Succeed())
		Expect(t, v.Validate(10), Succeed())
		Expect(t, v.Validate(99), Succeed())
		Expect(t, v.Validate(100), Succeed())
	})

	t.Run("Builder", func(t *testing.T) {
		// r, v := rule.NewBuilder("testing"), &va.RangeVa[int]{Min: new(10)}
		b := rule.NewBuilder("testing")
		r := must.NoErrorV(rule.Compile([]byte(`@testing[10,]`)))
		v := must.NoErrorV(va.NewRangeVa[int](r))

		v.BuiltTo(b)
		Expect(t, b.Bytes(), Equal([]byte(`@testing[10,]`)))

		b = rule.NewBuilder("testing")
		v = nil
		v.BuiltTo(b)
		Expect(t, b.Bytes(), Equal([]byte(`@testing`)))

		b = rule.NewBuilder("testing")
		r = must.NoErrorV(rule.Compile([]byte(`@testing[10,100)`)))
		v = must.NoErrorV(va.NewRangeVa[int](r))
		v.BuiltTo(b)
		Expect(t, b.Bytes(), Equal([]byte(`@testing[10,100)`)))
	})
}
