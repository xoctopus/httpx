package httpx_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/routex"
)

// Code is a codex code declaring status, identifier and hint
type Code int8

const CODE__NOT_FOUND Code = 1

func (Code) StatusCode() int        { return http.StatusNotFound }
func (Code) CodeIdentifier() string { return "TEST::NOT_FOUND" }
func (Code) CodeHint() string       { return "不存在" }

type Data struct {
	Name string `json:"name"`
}

// Case returns according to the query parameter name
type Case struct {
	httpx.MethodGet `path:"/case"`
	Name            string `in:"query" name:"name"`
}

func (r *Case) Output(context.Context) (any, error) {
	switch r.Name {
	case "nil":
		return nil, nil
	case "data":
		return &Data{Name: "x"}, nil
	case "wrap-status":
		return httpx.WrapResponse[any](nil, httpx.WithResponseStatusCode(http.StatusNoContent)), nil
	case "wrap-cookie":
		return httpx.WrapResponse(&Data{}, httpx.WithResponseCookies(&http.Cookie{Name: "c", Value: "v"})), nil
	case "wrap-content-type":
		return httpx.WrapResponse(&Data{Name: "x"}, httpx.WithResponseContentType("application/vnd.x+json")), nil
	case "error-options":
		return httpx.ErrorResponse(
			errors.New("PermissionDenied"),
			httpx.WithResponseStatusCode(http.StatusForbidden),
			httpx.WithResponseTitle("无权操作"),
			httpx.WithResponseDetail("需要管理员权限操作"),
		), nil
	case "error-override-status":
		return httpx.ErrorResponse(codex.New(CODE__NOT_FOUND), httpx.WithResponseStatusCode(http.StatusGone)), nil
	case "error-metadata":
		return httpx.ErrorResponse(errors.New("x"), httpx.WithResponseMetadata("X-Reason", "a")), nil
	case "error-content-type":
		return httpx.ErrorResponse(codex.New(CODE__NOT_FOUND), httpx.WithResponseContentType("application/json")), nil
	default:
		return nil, codex.New(CODE__NOT_FOUND)
	}
}

type CreateCase struct {
	httpx.MethodPost `path:"/case"`
}

func (CreateCase) Output(context.Context) (any, error) { return &Data{Name: "x"}, nil }

var executed bool

// Required has a required query parameter
type Required struct {
	httpx.MethodGet `path:"/required"`
	V               string `in:"query" name:"v"`
}

func (Required) Output(context.Context) (any, error) {
	executed = true
	return nil, nil
}

// Fail is a pre operator returning a codex error
type Fail struct{}

func (Fail) Output(context.Context) (any, error) { return nil, codex.New(CODE__NOT_FOUND) }

type Next struct {
	httpx.MethodGet `path:"/next"`
}

func (Next) Output(context.Context) (any, error) {
	executed = true
	return nil, nil
}

var handler = must.NoErrorV(routex.NewHandler(
	httpx.GroupRouter("/root").With(
		httpx.NewRouter(&Case{}),
		httpx.NewRouter(&CreateCase{}),
		httpx.NewRouter(&Required{}),
		httpx.NewRouter(&Fail{}, &Next{}),
	),
	"srv@v0.0.1",
))

type served struct {
	*httptest.ResponseRecorder
	body map[string]any
}

func serve(method, target string) *served {
	executed = false
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, httptest.NewRequest(method, "/root"+target, nil))
	r := &served{ResponseRecorder: rw}
	_ = json.Unmarshal(rw.Body.Bytes(), &r.body)
	return r
}

func (r *served) has(key string) bool { _, ok := r.body[key]; return ok }

// TestWrapResponse covers spec: 成功响应的状态码, 响应头与 cookie, Content-Type
func TestWrapResponse(t *testing.T) {
	bdd.From(t).Given("success responses", func(t bdd.T) {
		t.When("nil, nil", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=nil")
			t.Then("204", bdd.Equal(r.Code, http.StatusNoContent))
			t.Then("empty body", bdd.Equal(r.Body.Len(), 0))
		})
		t.When("GET data, nil", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=data")
			t.Then("200", bdd.Equal(r.Code, http.StatusOK))
			t.Then("json", bdd.HavePrefix(r.Header().Get("Content-Type"), "application/json"))
			t.Then("serialized value", bdd.Equal(r.body["name"], any("x")))
		})
		t.When("POST data, nil", func(t bdd.T) {
			t.Then("201", bdd.Equal(serve(http.MethodPost, "/case").Code, http.StatusCreated))
		})
		t.When("WrapResponse declares status", func(t bdd.T) {
			t.Then("204", bdd.Equal(serve(http.MethodGet, "/case?name=wrap-status").Code, http.StatusNoContent))
		})
		t.When("WrapResponse sets cookie", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=wrap-cookie")
			t.Then("cookie", bdd.ContainsSubString(r.Header().Get("Set-Cookie"), "c=v"))
		})
		t.When("WrapResponse overrides content type", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=wrap-content-type")
			t.Then("header", bdd.Equal(r.Header().Get("Content-Type"), "application/vnd.x+json"))
			t.Then("still json", bdd.Equal(r.body["name"], any("x")))
		})
	})
}

// TestErrorResponse covers spec: 响应选项, 错误响应的媒体类型, 响应头与 cookie
func TestErrorResponse(t *testing.T) {
	bdd.From(t).Given("error response options", func(t bdd.T) {
		t.When("plain error with problem content", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=error-options")
			t.Then("403", bdd.Equal(r.Code, http.StatusForbidden))
			t.Then("title", bdd.Equal(r.body["title"], any("无权操作")))
			t.Then("detail", bdd.Equal(r.body["detail"], any("需要管理员权限操作")))
			t.Then("no code", bdd.BeFalse(r.has("code")))
			t.Then("no key", bdd.BeFalse(r.has("key")))
			t.Then("text not leaked", bdd.BeFalse(strings.Contains(r.Body.String(), "PermissionDenied")))
		})
		t.When("status declared by codex is overridden", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=error-override-status")
			t.Then("410", bdd.Equal(r.Code, http.StatusGone))
			t.Then("key kept", bdd.Equal(r.body["key"], any("TEST::NOT_FOUND")))
		})
		t.When("response header set", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=error-metadata")
			t.Then("header", bdd.Equal(r.Header().Get("X-Reason"), "a"))
		})
	})

	bdd.From(t).Given("error media type", func(t bdd.T) {
		t.When("default", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=error")
			t.Then("problem+json", bdd.Equal(r.Header().Get("Content-Type"), "application/problem+json"))
			t.Then("404", bdd.Equal(r.Code, http.StatusNotFound))
		})
		t.When("overridden", func(t bdd.T) {
			r := serve(http.MethodGet, "/case?name=error-content-type")
			t.Then("json", bdd.Equal(r.Header().Get("Content-Type"), "application/json"))
			t.Then("body unchanged", bdd.Equal(r.body["key"], any("TEST::NOT_FOUND")))
		})
	})
}

// TestErrorExit covers spec: 错误的统一出口
func TestErrorExit(t *testing.T) {
	bdd.From(t).Given("request parameter error", func(t bdd.T) {
		t.When("required query is absent", func(t bdd.T) {
			r := serve(http.MethodGet, "/required")
			t.Then("400", bdd.Equal(r.Code, http.StatusBadRequest))
			t.Then("problem+json", bdd.Equal(r.Header().Get("Content-Type"), "application/problem+json"))
			t.Then("operator not executed", bdd.BeFalse(executed))
		})
	})

	bdd.From(t).Given("pre operator returns an error", func(t bdd.T) {
		t.When("request is served", func(t bdd.T) {
			r := serve(http.MethodGet, "/next")
			t.Then("404 from codex", bdd.Equal(r.Code, http.StatusNotFound))
			t.Then("key", bdd.Equal(r.body["key"], any("TEST::NOT_FOUND")))
			t.Then("following operator not executed", bdd.BeFalse(executed))
		})
	})
}

func ExampleErrorResponse() {
	r := serve(http.MethodGet, "/case?name=error-options")
	fmt.Println(r.Code, r.Header().Get("Content-Type"))
	fmt.Println(strings.TrimSpace(r.Body.String()))

	// Output:
	// 403 application/problem+json
	// {"title":"无权操作","detail":"需要管理员权限操作","errors":[{"src":"srv@v0.0.1"}]}
}
