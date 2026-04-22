package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/xoctopus/httpx/internal/payload"
	"github.com/xoctopus/httpx/internal/payload/path"
)

type Request interface {
	Context() context.Context
	Underlying() *http.Request

	Method() string

	Path() string
	Param(string) string

	Header() http.Header
	HeaderValue(string) string
	HeaderValues(string) []string

	Query() string
	QueryValue(string) string
	QueryValues(string) []string

	Cookies() []*http.Cookie
	CookieValue(string) string
	CookieValues(string) []string

	ValueIn(string, string) string
	ValuesIn(string, string) []string

	Body() io.ReadCloser
}

func From(r *http.Request) Request {
	return &request{
		request:   r,
		timestamp: time.Now(),
	}
}

type request struct {
	request   *http.Request
	timestamp time.Time
	queries   url.Values
	cookies   []*http.Cookie
	params    path.ValueGetter
}

func (r *request) Context() context.Context {
	return r.request.Context()
}

func (r *request) Underlying() *http.Request {
	return r.request
}

func (r *request) Method() string {
	return r.request.Method
}

func (r *request) Path() string {
	return r.request.URL.Path
}

func (r *request) Param(name string) string {
	if r.params == nil {
		r.params = path.ParamGetterFrom(r.Context())
	}
	v := r.params.PathValue(name)
	if len(v) > 0 {
		if p, err := url.PathUnescape(v); err == nil {
			return p
		}
	}
	return ""
}

func (r *request) Header() http.Header {
	return r.request.Header
}

func (r *request) HeaderValue(name string) string {
	if vs := r.HeaderValues(name); len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (r *request) HeaderValues(name string) []string {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if vs := r.QueryValues("x-param-header-" + name); len(vs) > 0 {
		return vs
	}
	if vs, ok := r.request.Header[name]; ok {
		return vs
	}
	return nil
}

func (r *request) Query() string {
	return r.request.URL.RawQuery
}

func (r *request) QueryValue(name string) string {
	if vs := r.QueryValues(name); len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (r *request) QueryValues(name string) []string {
	if r.queries == nil {
		r.queries = r.request.URL.Query()
	}
	if len(r.queries) == 0 {
		if strings.HasPrefix(
			r.HeaderValue("Content-Type"),
			"application/x-www-form-urlencoded",
		) {
			data, err := io.ReadAll(r.request.Body)
			if err == nil {
				_ = r.request.Body.Close()
				if q, e := url.ParseQuery(string(data)); e == nil {
					r.queries = q
				}
				// put back to body for custom parse
				r.request.Body = io.NopCloser(bytes.NewBuffer(data))
			}
		}
	}
	return r.queries[name]
}

func (r *request) Cookies() []*http.Cookie {
	return r.cookies
}

func (r *request) CookieValue(name string) string {
	if vs := r.CookieValues(name); len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (r *request) CookieValues(name string) []string {
	values := make([]string, 0)
	for _, c := range r.cookies {
		if c.Name == name {
			if c.Expires.IsZero() {
				values = append(values, c.Value)
			} else if c.Expires.After(r.timestamp) {
				values = append(values, c.Value)
			}
		}
	}
	return values
}

func (r *request) ValueIn(loc, key string) string {
	values := r.ValuesIn(loc, key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (r *request) ValuesIn(loc, key string) []string {
	switch loc {
	case payload.HEADER:
		return r.QueryValues(key)
	case payload.PATH:
		p := r.Param(key)
		if len(p) == 0 {
			return nil
		}
		v, _ := url.PathUnescape(p)
		if len(v) == 0 {
			return nil
		}
		return []string{v}
	case payload.QUERY:
		return r.QueryValues(key)
	case payload.COOKIE:
		return r.CookieValues(key)
	default:
		return nil
	}
}

func (r *request) Body() io.ReadCloser {
	rc := r.request.Body
	if r.request.ContentLength == 0 {
		if v := r.request.Header.Get("Content-Type"); len(v) == 0 {
			if q := r.request.URL.RawQuery; len(q) > 0 {
				r.request.Header.Set("Content-Type", `application/x-www-form-urlencoded; param="value"`)
				rc = io.NopCloser(bytes.NewBufferString(q))
			}
		}
	}
	return ReadCloserWithHeader(rc, r.request.Header)
}

func ReadCloserWithHeader(rc io.ReadCloser, header http.Header) io.ReadCloser {
	return &WithHeaderReadCloser{
		ReadCloser: rc,
		header:     header,
	}
}

type WithHeaderReadCloser struct {
	header http.Header
	io.ReadCloser
}

func (b *WithHeaderReadCloser) Header() http.Header {
	return b.header
}
