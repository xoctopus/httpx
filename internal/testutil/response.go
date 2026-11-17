package testutil

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"reflect"

	"github.com/xoctopus/x/testx"
)

func BeResponse[E ~string | ~[]byte | *http.Response](e E) testx.Matcher[*http.Response] {
	return &responseMatcher{
		expect: UnifyResponse(e),
	}
}

type responseMatcher struct {
	expect string
	actual string
}

func (m *responseMatcher) Negative() bool {
	return false
}

func (m *responseMatcher) Action() string {
	return "Be response"
}

func (m *responseMatcher) Match(r *http.Response) bool {
	m.actual = UnifyResponse(r)
	return m.actual == m.expect
}

func (m *responseMatcher) NormalizeExpect() any {
	return m.expect
}

func (m *responseMatcher) NormalizeActual(r *http.Response) any {
	return m.actual
}

func UnifyResponse[E ~string | ~[]byte | *http.Response](rsp E) string {
	var data []byte

	switch reflect.TypeOf(rsp).Kind() {
	case reflect.String:
		data = []byte(any(rsp).(string))
	case reflect.Slice:
		data = any(rsp).([]byte)
	default:
		// dump a copy and give the body back, so the caller can still read it;
		// the known length also stops DumpResponse adding `Connection: close`
		x := any(rsp).(*http.Response)
		r := *x
		if r.Body != nil {
			body, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			x.Body = io.NopCloser(bytes.NewReader(body))
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
		}
		data, _ = httputil.DumpResponse(&r, true)
	}

	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)
	data = bytes.TrimSpace(data)
	return string(data)
}
