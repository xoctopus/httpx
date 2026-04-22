package va_test

import (
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func TestRegexpVa(t *testing.T) {
	r := must.NoErrorV(rule.Compile([]byte(`@string/^[a-zA-Z]+$/`)))
	v := va.NewRegexpVa(r, "非空字符串, 仅包含a-zA-Z")

	err := v.Validate("123")
	t.Log(err)
	Expect(t, err, IsCodeError(va.ERROR__NOT_MATCH_REGEXP))
	Expect(t, v.Validate("abc"), Succeed())

	b := rule.NewBuilder("string")
	v.BuiltTo(b)
	Expect(t, b.Bytes(), Equal(r.Bytes()))

	b = rule.NewBuilder("string")
	v = va.NewRegexpVa(b, "nothing")
	Expect(t, v, BeNil[*va.RegexpVa]())
}
