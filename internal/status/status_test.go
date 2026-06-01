package status_test

import (
	"errors"
	"net/http"
	"slices"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/validation"
)

func TestErrorsFrom(t *testing.T) {
	errs := slices.Collect(status.ErrorsFrom(nil))
	Expect(t, errs, HaveLen[[]error](0))

	err0 := errors.New("0")

	err1 := validation.WrapPositionError(err0, "0")
	errs = slices.Collect(status.ErrorsFrom(err1))
	Expect(t, errs, HaveLen[[]error](1))
	Expect(t, errs[0], Equal(err1))

	err2 := status.Wrap(err0, status.AsStatus(http.StatusBadRequest))
	errs = slices.Collect(status.ErrorsFrom(err2))
	Expect(t, errs, HaveLen[[]error](1))
	Expect(t, errs[0], Equal(err2))

	err3 := validation.WrapLocationError(err0, "3")
	errs = slices.Collect(status.ErrorsFrom(err3))
	Expect(t, errs, HaveLen[[]error](2))
	Expect(t, errs[0], Equal(err3))
	Expect(t, errs[1], Equal(err0))

	err4 := errors.Join(err0, err1, nil, err2, err3)
	errs = slices.Collect(status.ErrorsFrom(err4))
	Expect(t, errs, HaveLen[[]error](5))
}
