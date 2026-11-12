package status

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
)

// AsDetail resolves err as a single problem: the status is the first declared
// one from outside in, and code, key, title and detail come from the outermost
// codex error (the identity layer) only. a downstream *Detail is returned as is.
func AsDetail(err error, src string) *Detail {
	if x, ok := errors.AsType[*Detail](err); ok {
		return x
	}
	_, entries := resolve(err, resolution{}, src)
	if len(entries) > 0 {
		return entries[0]
	}
	return (&resolution{}).asDetail(err.Error(), src)
}

// UnmarshalResponse parses an error body from a downstream service.
// The body is expected to be a Response (RFC 9457 style problem details).
// msg and message are compatible with the common {code, msg} style error body.
// A non JSON body is kept in Detail and the status code turns to 500.
func UnmarshalResponse(code int, rspraw []byte) *Detail {
	d := &Detail{Status: code}

	if len(rspraw) > 0 {
		rsp := &struct {
			Response `json:",inline"`
			Msg      string `json:"msg,omitzero"`
			Message  string `json:"message,omitzero"`
		}{}
		if err := json.Unmarshal(rspraw, rsp); err != nil {
			d.Status = http.StatusInternalServerError
			d.Detail = string(rspraw)
		} else {
			d.Code = rsp.Code
			d.Key = rsp.Key
			d.Title = cmp.Or(rsp.Title, rsp.Msg, rsp.Message)
			d.Detail = rsp.Detail
			d.Errors = rsp.Errors
		}
	}

	d.Key = cmp.Or(d.Key, d.StatusText())
	d.Title = cmp.Or(d.Title, http.StatusText(d.Status))
	return d
}

// Detail is an item of problem details. json fields follow RFC 9457 where the
// semantics match; the others are extension members.
type Detail struct {
	// Code biz error code
	Code int64 `json:"code,omitzero"`
	// Key code identifier
	Key string `json:"key,omitzero"`
	// Title code hint, the same for every occurrence of the problem
	Title string `json:"title,omitzero"`
	// Detail explanation of this occurrence, for client not for debugging
	Detail string `json:"detail,omitzero"`
	// Pointer JSON Pointer (RFC 6901) to the error position in the request
	// document. the document roots are payload.Locations, eg: `/body/items/0/age`,
	// `/query/ids/1`, `/header/X-Request-Id`. the pointer of a downstream Detail
	// is relative to the downstream request.
	Pointer jsontext.Pointer `json:"pointer,omitzero"`
	// Src caused source. eg: srv-abc@v0.0.1
	Src string `json:"src,omitzero"`
	// Desc the full text of the error, for logging only, never presented
	Desc string `json:"-"`
	// Errors sub problems from downstream
	Errors []*Detail `json:"errors,omitzero"`

	// Status http status code, 0 means unspecified (eg: downstream sub problems)
	Status int `json:"-"`
}

func (d *Detail) StatusCode() int {
	return d.Status
}

func (d *Detail) StatusText() string {
	return AsStatus(d.Status).StatusText()
}

func (d *Detail) SetDetail(detail string) {
	d.Detail = detail
}

// Error is for logging: the key, or the status text when there is no key, with
// the first of detail, title and desc.
func (d *Detail) Error() string {
	return fmt.Sprintf("%s{%s}", cmp.Or(d.Key, d.StatusText()), cmp.Or(d.Detail, d.Title, d.Desc))
}

// UnmarshalResponse keeps Src if it is set by caller, eg: downstream host.
func (d *Detail) UnmarshalResponse(code int, data []byte) error {
	src := d.Src
	*d = *UnmarshalResponse(code, data)
	d.Src = cmp.Or(d.Src, src)
	return nil
}
