package status

import (
	"net/http"
	"strconv"

	"github.com/xoctopus/x/slicex"
)

func AsResponse(err error, source string) *Response {
	if err == nil {
		return nil
	}

	var (
		er       *Response
		location = ""
	)

	for e := range ErrorsFrom(err) {
		if x, ok := e.(HasLocation); ok {
			location = x.Location()
			continue
		}

		ee := AsDescription(e, source, location)
		if er == nil {
			er = &Response{
				Code:    ee.Status.StatusCode(),
				Message: ee.Message,
			}
			if ee.Status != nil && ee.Status.StatusCode() == http.StatusBadRequest {
				er.Message = ee.Status.StatusText()
			}
		}

		if len(ee.Errors) > 0 {
			er.Errors = append(er.Errors, ee.Errors...)
		} else {
			er.Errors = append(er.Errors, ee)
		}
	}

	if er == nil {
		return &Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return er
}

type Response struct {
	// 错误码
	Code int `json:"code,omitzero"`
	// 错误信息
	Message string `json:"message,omitzero"`
	// 错误详情
	Errors []*Description `json:"errors,omitzero"`

	Extra map[string]any `json:",inline"`
}

// StatusCode returns http.StatusCode
// For extending the semantics of HTTP status codes, the original HTTP status codes
// are multiplied to represent specific business logic errors.
// For example:
// http.StatusBadRequest = 400
// 400000001: Missing field A
// 400000002: Invalid field B
func (e *Response) StatusCode() int {
	if e.Code > 999 {
		i, _ := strconv.ParseUint(strconv.FormatUint(uint64(e.Code), 10)[0:3], 10, 64)
		return int(i)
	}
	return e.Code
}

func (e *Response) Unwrap() []error {
	if e.Errors != nil {
		return slicex.M(e.Errors, func(e *Description) error { return e })
	}
	return nil
}
