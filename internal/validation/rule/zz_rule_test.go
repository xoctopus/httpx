package rule_test

import (
	"testing"

	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/rule"
)

func TestNewRule(t *testing.T) {
	for _, text := range []string{
		`@string<byte>[10]`,
		`@string<byte>[10,16]`,
		`@string<byte>(10,16)`, // invalid range
		`@string<byte>[3]{abc,def,ghi}`,
		`@string<byte>[3]{abc,def,ghi}/\w+\/abc/`,
		`@string?`,
		`@string='abc'`,
		`@slice<@float64<22,4>>`,
	} {
		r, err := rule.Compile(text)
		testx.Expect(t, err, testx.Succeed())
		testx.Expect(t, r.Bytes(), testx.Equal([]byte(text)))
	}
}
