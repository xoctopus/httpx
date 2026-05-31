package status

import (
	"fmt"
	"net/http"

	"github.com/xoctopus/x/stringsx"
)

func AsStatus(code int) Status {
	return &_status{
		code: code,
		text: stringsx.UpperSnakeCase(http.StatusText(code)),
	}
}

type _status struct {
	code int
	text string
}

func (s _status) IsValid() bool {
	return s.code > 0
}

func (s _status) StatusCode() int {
	return s.code
}

func (s _status) StatusText() string {
	return s.text
}

func Wrap(err error, s Status) error {
	return &Err{code: s.StatusCode(), text: s.StatusText(), err: err}
}

func WrapCode(err error, code int) error {
	return Wrap(err, AsStatus(code))
}

type Err struct {
	code int
	text string
	err  error
}

func (e *Err) StatusCode() int {
	return e.code
}

func (e *Err) StatusText() string {
	return e.text
}

func (e *Err) Error() string {
	return fmt.Sprintf("%s{message=%q,status=%d}", e.text, e.err, e.code)
}

func (e *Err) Unwrap() error {
	return e.err
}
