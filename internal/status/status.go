package status

import (
	"errors"
	"fmt"
	"net/http"
)

type Describer interface {
	StatusCode() int
}

type HasStatusText interface {
	Status() string
}

type Status interface {
	Describer
	HasStatusText
}

type Modifier interface {
	SetStatusCode(int)
}

type HasPosition interface {
	Position() string
}

type HasLocation interface {
	Location() string
}

func Wrap(err error, code int, text string) error {
	if err == nil {
		return nil
	}
	if len(text) == 0 {
		text = http.StatusText(code)
	}

	return &_StatusError{
		text: text,
		code: code,
		err:  err,
	}
}

func WrapStatus(err error, status Status) error {
	return Wrap(err, status.StatusCode(), status.Status())
}

type _StatusError struct {
	code int
	text string
	err  error
}

func (e *_StatusError) Status() string {
	return e.text
}

func (e *_StatusError) StatusCode() int {
	return e.code
}

func (e *_StatusError) Error() string {
	return fmt.Sprintf("%s{message=%q,status=%d}", e.text, e.err, e.code)
}

func (e *_StatusError) Unwrap() error {
	return e.err
}

type Description struct {
	// Code error code
	Code string `json:"code,omitzero"`
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

	Status int `json:"-"`
}

func (d *Description) StatusCode() int {
	return d.Status
}

func (d *Description) Error() string {
	return fmt.Sprintf("%s{message=%q}", d.Code, d.Message)
}

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
		de.Status = x.StatusCode()
	}

	if x, ok := err.(HasStatusText); ok {
		de.Code = x.Status()
	}

	if x, ok := err.(HasPosition); ok {
		de.Code = ""
		de.Status = http.StatusBadRequest
		de.Position = x.Position()
	}

	if de.Code == "" {
		de.Code = http.StatusText(http.StatusInternalServerError)
	}

	if de.Status == 0 {
		de.Status = http.StatusInternalServerError
	}

	return de
}
