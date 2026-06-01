package request_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/payload/path"
	"github.com/xoctopus/httpx/internal/request"
)

var (
	srv *http.Server
	req *http.Request
)

func init() {
	go func() {
		mux := &http.ServeMux{}
		mux.HandleFunc("PUT /v1/user/{userID}", func(_ http.ResponseWriter, r *http.Request) {
			r.Header.Add("h1", "v1")
			r.Header.Add("h1", "v2")
			r.AddCookie(&http.Cookie{Name: "token", Value: "1"})
			r.AddCookie(&http.Cookie{Name: "token", Value: "2"})
			req = r.WithContext(path.WithParamGetter(r.Context(), r))

			data, _ := io.ReadAll(r.Body)
			defer func() { _ = r.Body.Close() }()
			req.Body = io.NopCloser(bytes.NewReader(data))
		})
		srv = &http.Server{
			Addr:    "0.0.0.0:9999",
			Handler: mux,
		}

		_ = srv.ListenAndServe()
	}()
	time.Sleep(time.Second)
}

func TestRequest(t *testing.T) {
	defer func() { _ = srv.Close() }()

	req, _ = http.NewRequest(
		"PUT",
		"http://localhost:9999/v1/user/100?q1=v1&q1=v2&x-param-header-H2=v1",
		bytes.NewBuffer([]byte(`body`)),
	)
	_, _ = (&http.Client{}).Do(req)

	r := request.From(req)
	Expect(t, r.Underlying(), Equal(req))
	Expect(t, r.Method(), Equal("PUT"))

	Expect(t, r.Path(), Equal("/v1/user/100"))
	Expect(t, r.PathParam("userID"), Equal("100"))
	Expect(t, r.PathParam("other"), Equal(""))

	Expect(t, r.Query(), Equal("q1=v1&q1=v2&x-param-header-H2=v1"))
	Expect(t, r.QueryValue("q1"), Equal("v1"))
	Expect(t, r.QueryValues("q1"), Equal([]string{"v1", "v2"}))
	Expect(t, r.QueryValue("q2"), Equal(""))

	Expect(t, r.Header().Values("h1"), Equal([]string{"v1", "v2"}))
	Expect(t, r.HeaderValue("h1"), Equal("v1"))
	Expect(t, r.HeaderValues("h1"), Equal([]string{"v1", "v2"}))
	Expect(t, r.HeaderValue("h2"), Equal("v1"))
	Expect(t, r.HeaderValue("h3"), Equal(""))

	Expect(t, r.Cookies(), HaveLen[[]*http.Cookie](2))
	Expect(t, r.CookieValue("token"), Equal("1"))
	Expect(t, r.CookieValue("other"), Equal(""))
	Expect(t, r.CookieValues("token"), Equal([]string{"1", "2"}))

	Expect(t, r.ValuesIn("header", "h1"), Equal(r.HeaderValues("h1")))
	Expect(t, r.ValuesIn("query", "q1"), Equal(r.QueryValues("q1")))
	Expect(t, r.ValuesIn("cookie", "token"), Equal(r.CookieValues("token")))
	Expect(t, r.ValuesIn("path", "userID"), Equal([]string{r.PathParam("userID")}))
	Expect(t, r.ValuesIn("path", "other"), HaveLen[[]string](0))
	Expect(t, r.ValueIn("other", "key"), Equal(""))
	Expect(t, r.ValueIn("path", "userID"), Equal("100"))

	body := r.Body()
	defer func() { _ = body.Close() }()
	data, err := io.ReadAll(body)
	Expect(t, err, Succeed())
	Expect(t, data, Equal([]byte("body")))

	t.Run("QueryInURLEncodedBody", func(t *testing.T) {
		req, _ = http.NewRequest(
			"PUT", "/",
			bytes.NewBuffer([]byte(`q2=v1&q2=v2`)),
		)
		req.Header.Add("content-type", "application/x-www-form-urlencoded")

		r = request.From(req)
		Expect(t, r.QueryValues("q2"), Equal([]string{"v1", "v2"}))
	})

	t.Run("DefaultToURLEncodedBody", func(t *testing.T) {
		req, _ = http.NewRequest("PUT", "/root?q3=v1", nil)
		req.Header.Add("user-defined", "my-value")
		r = request.From(req)

		body2 := r.Body()
		defer func() { _ = body2.Close() }()
		data, err = io.ReadAll(body2)
		Expect(t, err, Succeed())
		Expect(t, data, Equal([]byte("q3=v1")))

		header := body2.(content.WithHeader).Header()
		Expect(t, header.Get("user-defined"), Equal("my-value"))
	})
}
