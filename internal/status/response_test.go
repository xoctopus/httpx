package status_test

import (
	"errors"
	"net/http"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/validation"
)

func TestErrorResponse(t *testing.T) {
	r := status.AsResponse(nil, "any")
	Expect(t, r, BeNil[*status.Response]())

	err0 := errors.New("0")
	err1 := validation.WrapPositionError(err0, "pos")
	err2 := validation.WrapLocationError(err1, "loc")
	r = status.AsResponse(err2, "srv-any")
	Expect(t, r.Errors, HaveLen[[]*status.Description](1))
	Expect(t, r.Errors[0].Location, Equal("loc"))
	Expect(t, r.Errors[0].Position, Equal("pos"))
	Expect(t, r.Errors[0].Message, Equal("0"))
	Expect(t, r.Unwrap(), HaveLen[[]error](1))

	r1 := &status.Response{Code: http.StatusForbidden * 1e6}
	Expect(t, r1.StatusCode(), Equal(http.StatusForbidden))
	r2 := &status.Response{Code: http.StatusForbidden}
	Expect(t, r2.StatusCode(), Equal(http.StatusForbidden))
	Expect(t, r2.Unwrap(), HaveLen[[]error](0))
}
