package status

import (
	"fmt"
	"net/http"

	"github.com/xoctopus/x/stringsx"
)

func AsStatus(code int) Status {
	s := &_status{code: code}
	// non-standard code (eg: 0, 499, 520) has no status text
	if text := http.StatusText(code); text != "" {
		s.text = stringsx.UpperSnakeCase(text)
	}
	return s
}

type _status struct {
	code int
	text string
}

func (s *_status) IsValid() bool {
	return s != nil && s.code > 0
}

func (s *_status) StatusCode() int {
	return s.code
}

func (s *_status) StatusText() string {
	return s.text
}

func Wrap(err error, s Status) error {
	return WrapStatus(err, s.StatusCode(), s.StatusText())
}

func WrapCode(err error, code int) error {
	return Wrap(err, AsStatus(code))
}

func WrapStatus(err error, code int, text string) error {
	if err == nil {
		return nil
	}
	if code >= http.StatusOK && code < http.StatusMultipleChoices {
		// should panic here?
		return nil
	}
	return &Err{code: code, text: text, err: err}
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
