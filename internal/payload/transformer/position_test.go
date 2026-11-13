package transformer_test

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/payload/transformer"
	"github.com/xoctopus/httpx/internal/validation"
)

// position is the error position of a request: location is one of payload.Locations
// and pointer is a JSON Pointer relative to the root of the location
type position struct {
	location string
	pointer  jsontext.Pointer
}

// TestPosition asserts the position of validation errors in a request,
// spec: 错误位置, 校验的失败顺序
func TestPosition(t *testing.T) {
	pointer := func(t *testing.T, req *http.Request, v any) position {
		err := transformer.UnmarshalUnderlyingRequest(req, v)
		x, ok := errors.AsType[*validation.PointerError](err)
		Expect(t, ok, BeTrue())
		return position{x.Location(), x.Pointer()}
	}
	at := func(location string, pointer jsontext.Pointer) Matcher[position] {
		return Equal(position{location, pointer})
	}

	t.Run("Path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue("id", "x")
		v := &struct {
			ID int `in:"path" name:"id"`
		}{}
		Expect(t, pointer(t, req, v), at("path", "/id"))
	})
	t.Run("QuerySingle", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?size=x", nil)
		v := &struct {
			Size int `in:"query" name:"size"`
		}{}
		Expect(t, pointer(t, req, v), at("query", "/size"))
	})
	t.Run("QueryMultiple", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?ids=1&ids=x", nil)
		v := &struct {
			IDs []int `in:"query" name:"ids"`
		}{}
		Expect(t, pointer(t, req, v), at("query", "/ids/1"))
	})
	t.Run("QueryMissingRequired", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		v := &struct {
			Size int `in:"query" name:"size"`
		}{}
		Expect(t, pointer(t, req, v), at("query", "/size"))
	})
	t.Run("HeaderCanonical", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Create-Time", "x")
		v := &struct {
			CreateTime int `in:"header" name:"x-create-time"`
		}{}
		Expect(t, pointer(t, req, v), at("header", "/X-Create-Time"))
	})
	t.Run("Cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "x"})
		v := &struct {
			Session int `in:"cookie" name:"session"`
		}{}
		Expect(t, pointer(t, req, v), at("cookie", "/session"))
	})
	t.Run("BodyJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"items":[{"age":1},{"age":"x"}]}`))
		req.Header.Set("Content-Type", "application/json")
		v := &struct {
			Body struct {
				Items []struct {
					Age int `json:"age"`
				} `json:"items"`
			} `in:"body" mime:"json"`
		}{}
		Expect(t, pointer(t, req, v), at("body", "/items/1/age"))
	})
	t.Run("BodyJSONArrayElement", func(t *testing.T) {
		for raw, want := range map[string]Matcher[position]{
			`{"ids":["x"]}`:   at("body", "/ids/0"),
			`{"ids":[1,"x"]}`: at("body", "/ids/1"),
		} {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			v := &struct {
				Body struct {
					IDs []int `json:"ids"`
				} `in:"body" mime:"json"`
			}{}
			Expect(t, pointer(t, req, v), want)
		}
	})
	t.Run("BodyURLEncoded", func(t *testing.T) {
		type Data struct {
			Age  int   `json:"age"`
			Tags []int `json:"tags"`
		}
		for raw, want := range map[string]Matcher[position]{
			"age=x":               at("body", "/age"),
			"age=1&tags=1&tags=x": at("body", "/tags/1"),
		} {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(raw))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			v := &struct {
				Data `in:"body" mime:"urlencoded"`
			}{}
			Expect(t, pointer(t, req, v), want)
		}
	})
	t.Run("BodyMultipart", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		w := multipart.NewWriter(buf)
		_ = w.WriteField("age", "1")
		_ = w.WriteField("tags", "1")
		_ = w.WriteField("tags", "x")
		_ = w.Close()
		req := httptest.NewRequest(http.MethodPost, "/", buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		v := &struct {
			Data struct {
				Age  int   `json:"age"`
				Tags []int `json:"tags"`
			} `in:"body" mime:"multipart"`
		}{}
		Expect(t, pointer(t, req, v), at("body", "/tags/1"))
	})
	t.Run("BodyInvalidJSON", func(t *testing.T) {
		// spec: 错误位置 / 请求体不是合法 JSON
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"age":`))
		req.Header.Set("Content-Type", "application/json")
		v := &struct {
			Body struct {
				Age int `json:"age"`
			} `in:"body" mime:"json"`
		}{}
		Expect(t, pointer(t, req, v), at("body", ""))
	})
	t.Run("HeaderBeforeBody", func(t *testing.T) {
		// spec: 校验的失败顺序 / 多个 location 同时有错
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"age":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Create-Time", "x")
		v := &struct {
			Body struct {
				Age int `json:"age"`
			} `in:"body" mime:"json"`
			CreateTime int `in:"header" name:"x-create-time"`
		}{}
		Expect(t, pointer(t, req, v), at("header", "/X-Create-Time"))
	})
	t.Run("QueryBeforeBody", func(t *testing.T) {
		// spec: 校验的失败顺序 / 多个 location 同时有错
		req := httptest.NewRequest(http.MethodPost, "/?size=x", strings.NewReader(`{"age":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		v := &struct {
			Body struct {
				Age int `json:"age"`
			} `in:"body" mime:"json"`
			Size int `in:"query" name:"size"`
		}{}
		Expect(t, pointer(t, req, v), at("query", "/size"))
	})
	t.Run("BodyDocumentOrderBeforeMissing", func(t *testing.T) {
		// spec: 校验的失败顺序 / 同一 body 内的多个错误
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"age":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		v := &struct {
			Body struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			} `in:"body" mime:"json"`
		}{}
		Expect(t, pointer(t, req, v), at("body", "/age"))
	})
}
