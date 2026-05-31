package status

import (
	"net/http"
	"strconv"

	"github.com/xoctopus/x/slicex"
)

func AsErrorResponse(err error, source string) *ErrorResponse {
	if err == nil {
		return nil
	}

	var er *ErrorResponse

	loc := ""

	for e := range ErrorsFrom(err) {
		if x, ok := e.(HasLocation); ok {
			loc = x.Location()
			continue
		}

		ee := AsDescription(e, source, loc)
		if er == nil {
			er = &ErrorResponse{
				Code: ee.Status.StatusCode(),
				Msg:  ee.Message,
			}
			if ee.Status.StatusCode() == http.StatusBadRequest {
				er.Msg = ee.Status.StatusText()
			}
		}

		if len(ee.Errors) > 0 {
			er.Errors = append(er.Errors, ee.Errors...)
		} else {
			er.Errors = append(er.Errors, ee)
		}
	}

	if er == nil {
		return &ErrorResponse{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}

	return er
}

type ErrorResponse struct {
	Code   int            `json:"code,omitzero"`
	Msg    string         `json:"msg,omitzero"`
	Errors []*Description `json:"errors,omitzero"`
	Extra  map[string]any `json:",inline"`
}

// StatusCode returns http.StatusCode
// For extending the semantics of HTTP status codes, the original HTTP status codes
// are multiplied by 1e6 to represent specific business logic errors.
// For example:
// http.StatusBadRequest = 400
// 400000001: Missing field A
// 400000002: Invalid field B
func (e *ErrorResponse) StatusCode() int {
	if e.Code > 1000 {
		i, _ := strconv.ParseUint(strconv.FormatUint(uint64(e.Code), 10)[0:3], 10, 64)
		return int(i)
	}
	return e.Code
}

func (e *ErrorResponse) Unwrap() []error {
	if e.Errors != nil {
		return slicex.M(e.Errors, func(e *Description) error { return e })
	}
	return nil
}
