package status

import (
	"net/http"

	"github.com/xoctopus/x/slicex"
)

// MediaType of error response, see RFC 9457
const MediaType = "application/problem+json"

// AsResponse resolves err as the root of problem details. the top level takes
// the identity layer above errors.Join if any, otherwise the first entry.
func AsResponse(err error, src string) *Response {
	if err == nil {
		return nil
	}

	top, entries := resolve(err, resolution{}, src)
	rsp := &Response{Errors: entries, Desc: err.Error()}

	switch {
	case top.ident || len(entries) == 0 || entries[0] == nil:
		rsp.Code, rsp.Key, rsp.Title, rsp.Detail = top.code, top.key, top.title, top.detail
	default:
		e := entries[0]
		rsp.Code, rsp.Key, rsp.Title, rsp.Detail = e.Code, e.Key, e.Title, e.Detail
	}

	rsp.status = top.status
	if rsp.status <= 0 && len(entries) > 0 {
		rsp.status = entries[0].Status
	}
	if rsp.status <= 0 {
		rsp.status = top.statusCode()
	}
	return rsp
}

// Response is the root of problem details. json fields follow RFC 9457 where
// the semantics match; the others are extension members. status and instance
// are not serialized: http status code is in status line, and the operation is
// identified by `Server` header.
type Response struct {
	// Code biz code
	Code int64 `json:"code,omitzero"`
	// Key code identifier
	Key string `json:"key,omitzero"`
	// Title code hint, the same for every occurrence of the problem
	Title string `json:"title,omitzero"`
	// Detail explanation of this occurrence, for client not for debugging
	Detail string `json:"detail,omitzero"`
	// Errors problems occurred in the same request
	Errors []*Detail `json:"errors,omitzero"`

	// Desc the full text of the error, for logging only, never presented
	Desc string `json:"-"`

	// status http status code
	status int
}

// SetStatusCode overrides status code, see Modifier
func (e *Response) SetStatusCode(code int) {
	e.status = code
}

func (e *Response) StatusCode() int {
	return e.status
}

func (e *Response) StatusText() string {
	return http.StatusText(e.status)
}

func (e *Response) SetTitle(title string) {
	e.Title = title
}

func (e *Response) SetDetail(detail string) {
	e.Detail = detail
}

func (e *Response) Unwrap() []error {
	if len(e.Errors) > 0 {
		return slicex.FilterMapping(
			e.Errors,
			func(e *Detail) (error, bool) {
				return e, e != nil
			},
		)
	}
	return nil
}
