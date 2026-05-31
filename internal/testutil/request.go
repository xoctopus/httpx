package testutil

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"reflect"
	"regexp"

	"github.com/xoctopus/x/testx"
)

func BeRequest[E ~string | ~[]byte | *http.Request](e E) testx.Matcher[*http.Request] {
	return &requestMatcher{
		expect: UnifyRequest(e),
	}
}

type requestMatcher struct {
	expect string
	actual string
}

func (m *requestMatcher) Negative() bool {
	return false
}

func (m *requestMatcher) Action() string {
	return "Be request"
}

func (m *requestMatcher) Match(r *http.Request) bool {
	m.actual = UnifyRequest(r)
	return m.actual == m.expect
}

func (m *requestMatcher) NormalizeExpect() any {
	return m.expect
}

func (m *requestMatcher) NormalizeActual(r *http.Request) any {
	return m.actual
}

var reContentTypeWithBoundary = regexp.MustCompile(`Content-Type: multipart/form-data; boundary=([A-Za-z0-9]+)`)

func UnifyRequest[E ~string | ~[]byte | *http.Request](req E) string {
	var data []byte

	switch reflect.TypeOf(req).Kind() {
	case reflect.String:
		data = []byte(any(req).(string))
	case reflect.Slice:
		data = any(req).([]byte)
	default:
		data, _ = httputil.DumpRequest(any(req).(*http.Request), true)
	}

	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)

	if reContentTypeWithBoundary.Match(data) {
		matches := reContentTypeWithBoundary.FindAllSubmatch(data, 1)
		data = bytes.Replace(data, matches[0][1], []byte("boundary1"), -1)
	}

	data = bytes.TrimSpace(data)
	return string(data)
}
