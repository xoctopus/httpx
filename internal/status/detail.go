package status

import (
	"errors"
	"fmt"
	"net/http"
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

type Description struct {
	// xxCode error code
	Text string `json:"code,omitzero"`
	// Message error message
	Message string `json:"message,omitzero"`
	// Detail error detail
	Detail string `json:"detail,omitzero"`
	// Location error location. enumerations in {query, header, path, body, cookie}
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
