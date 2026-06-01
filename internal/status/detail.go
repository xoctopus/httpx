package status

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
)

func AsDescription(err error, src, loc string) *Description {
	if x, ok := errors.AsType[*Description](err); ok {
		return x
	}

	de := &Description{
		Source:   src,
		Location: loc,
	}

	if x, ok := err.(interface{ Unwrap() error }); ok {
		if e := x.Unwrap(); e != nil {
			de.Message = e.Error()
		}
	}

	if len(de.Message) == 0 {
		de.Message = err.Error()
	}

	if x, ok := err.(Describer); ok {
		de.Status = AsStatus(x.StatusCode())
	}

	if x, ok := err.(HasCodeText); ok {
		de.Text = x.StatusText()
	}

	if x, ok := err.(HasLocation); ok {
		if de.Status == nil || !de.Status.IsValid() {
			de.Status = AsStatus(http.StatusBadRequest)
		}
		de.Location = x.Location()
	}

	if x, ok := err.(HasPosition); ok {
		if de.Status == nil || !de.Status.IsValid() {
			de.Status = AsStatus(http.StatusBadRequest)
		}
		de.Position = x.Position()
	}

	if de.Status == nil || !de.Status.IsValid() {
		de.Status = AsStatus(http.StatusInternalServerError)
	}

	if de.Text == "" {
		de.Text = de.Status.StatusText()
	}

	return de
}

func UnmarshalResponse(code int, rspraw []byte) *Description {
	d := &Description{
		Status: AsStatus(code),
	}

	defer func() {
		if len(d.Text) == 0 {
			d.Text = d.StatusText()
		}
		if len(d.Message) == 0 {
			d.Message = d.StatusText()
		}
	}()

	rsp := &Response{}
	if err := json.Unmarshal(rspraw, rsp); err != nil {
		d.Message = string(rspraw)
		d.Status = AsStatus(http.StatusInternalServerError)
		return d
	}

	messages := make([]string, 0, 4)
	switch len(rsp.Errors) {
	case 0:
		if len(rsp.Message) > 0 {
			messages = append(messages, rsp.Message)
		}
	case 1:
		d = rsp.Errors[0]
		if d.Status == nil {
			d.Status = AsStatus(code)
		}
	}

	// must be ordered: rsp.Message .. extra.title .. extra.detail
	if rsp.Extra != nil {
		if x, ok := rsp.Extra["title"].(string); ok && len(x) > 0 {
			messages = append(messages, x)
		}
		if x, ok := rsp.Extra["detail"].(string); ok && len(x) > 0 {
			messages = append(messages, x)
		}
	}
	if len(messages) > 0 {
		d.Message = strings.Join(messages, "\n")
	}
	return d
}

type Description struct {
	// Text error code text
	Text string `json:"code,omitzero"`
	// Message error message
	Message string `json:"message,omitzero"`
	// Detail error detail
	Detail string `json:"detail,omitzero"`
	// Location error location, see payload.Locations
	Location string `json:"location,omitzero"`
	// Position pointer to field positon. eg: Type.Field
	Position string `json:"position,omitzero"`
	// Source error caused source. eg: srv-abc@v1.1.0
	Source string `json:"source,omitzero"`
	// Errors error chain
	Errors []*Description `json:"errors,omitzero"`

	Status Status `json:"-"`
}

func (d *Description) StatusCode() int {
	return d.Status.StatusCode()
}

func (d *Description) StatusText() string {
	return d.Status.StatusText()
}

func (d *Description) Error() string {
	return fmt.Sprintf("%s{message=%q}", d.Text, d.Message)
}
