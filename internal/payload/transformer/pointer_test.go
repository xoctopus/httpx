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

// TestPointer asserts the JSON Pointer of validation errors in request document
func TestPointer(t *testing.T) {
	pointer := func(t *testing.T, req *http.Request, v any) jsontext.Pointer {
		err := transformer.UnmarshalUnderlyingRequest(req, v)
		x, ok := errors.AsType[*validation.PointerError](err)
		Expect(t, ok, BeTrue())
		return x.Pointer()
	}

	t.Run("Path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue("id", "x")
		v := &struct {
			ID int `in:"path" name:"id"`
		}{}
		Expect(t, pointer(t, req, v), Equal[jsontext.Pointer]("/path/id"))
	})
	t.Run("QuerySingle", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?size=x", nil)
		v := &struct {
			Size int `in:"query" name:"size"`
		}{}
		Expect(t, pointer(t, req, v), Equal[jsontext.Pointer]("/query/size"))
	})
	t.Run("QueryMultiple", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?ids=1&ids=x", nil)
		v := &struct {
			IDs []int `in:"query" name:"ids"`
		}{}
		Expect(t, pointer(t, req, v), Equal[jsontext.Pointer]("/query/ids/1"))
	})
	t.Run("QueryMissingRequired", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		v := &struct {
			Size int `in:"query" name:"size"`
		}{}
		Expect(t, pointer(t, req, v), Equal[jsontext.Pointer]("/query/size"))
	})
	t.Run("HeaderCanonical", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Create-Time", "x")
		v := &struct {
			CreateTime int `in:"header" name:"x-create-time"`
		}{}
		Expect(t, pointer(t, req, v), Equal[jsontext.Pointer]("/header/X-Create-Time"))
	})
	t.Run("Cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "x"})
		v := &struct {
			Session int `in:"cookie" name:"session"`
		}{}
		Expect(t, pointer(t, req, v), Equal[jsontext.Pointer]("/cookie/session"))
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
		Expect(t, pointer(t, req, v), Equal[jsontext.Pointer]("/body/items/1/age"))
	})
	t.Run("BodyJSONArrayElement", func(t *testing.T) {
		for raw, ptr := range map[string]jsontext.Pointer{
			`{"ids":["x"]}`:   "/body/ids/0",
			`{"ids":[1,"x"]}`: "/body/ids/1",
		} {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			v := &struct {
				Body struct {
					IDs []int `json:"ids"`
				} `in:"body" mime:"json"`
			}{}
			Expect(t, pointer(t, req, v), Equal(ptr))
		}
	})
	t.Run("BodyURLEncoded", func(t *testing.T) {
		type Data struct {
			Age  int   `json:"age"`
			Tags []int `json:"tags"`
		}
		for raw, ptr := range map[string]jsontext.Pointer{
			"age=x":               "/body/age",
			"age=1&tags=1&tags=x": "/body/tags/1",
		} {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(raw))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			v := &struct {
				Data `in:"body" mime:"urlencoded"`
			}{}
			Expect(t, pointer(t, req, v), Equal(ptr))
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
		Expect(t, pointer(t, req, v), Equal[jsontext.Pointer]("/body/tags/1"))
	})
}
