package scanner_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/scanner"
)

func TestWrapUnwrap(t *testing.T) {
	f := &scanner.Field{}
	f.Name = "stub"

	w := scanner.WrapField("stub_value", f)
	Expect(t, scanner.UnwrapField(w), Equal[any]("stub_value"))
	Expect(t, w.(scanner.WrappedField).Field(), Equal(f))
	Expect(t, scanner.UnwrapField(f), Equal[any](f))
}
