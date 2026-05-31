package validation_test

import (
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
	Expect(t, IsValidationError(&PositionError{}), BeTrue())
	Expect(t, IsValidationError(&LocationError{}), BeTrue())
	Expect(t, IsValidationError(nil), BeFalse())

}

func ExampleWrapLocationError() {
	wp := WrapPositionError(errors.New("wp"), "pos")
	fmt.Println(wp)

	wl := WrapLocationError(errors.New("wl"), "loc")
	fmt.Println(wl)

	fmt.Println(WrapLocationError(wp, "loc"))
	fmt.Println(WrapPositionError(wl, "pos"))

	fmt.Println(WrapPositionError(nil, ""))
	fmt.Println(WrapLocationError(nil, ""))

	// Output:
	// wp: [pos:pos]
	// wl: [loc:loc]
	// wp: [loc:loc pos:pos]
	// wl: [loc:loc pos:pos]
	// <nil>
	// <nil>
}
