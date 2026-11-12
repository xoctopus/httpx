package routex_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/routex"
)

var executed bool

// Deny is a pre operator returning an error response as its value
type Deny struct{}

func (Deny) Output(context.Context) (any, error) {
	return httpx.ErrorResponse(errors.New("PermissionDenied"), httpx.WithResponseStatusCode(http.StatusForbidden)), nil
}

type Allow struct{}

func (Allow) Output(context.Context) (any, error) { return nil, nil }

type Target struct {
	httpx.MethodGet `path:"/target"`
}

func (Target) Output(context.Context) (any, error) {
	executed = true
	return nil, nil
}

// TestNewHandler covers spec: 以返回值返回错误响应
func TestNewHandler(t *testing.T) {
	serve := func(pre httpx.Operator) *httptest.ResponseRecorder {
		executed = false
		root := httpx.GroupRouter("/root").With(httpx.NewRouter(pre, &Target{}))
		h := must.NoErrorV(routex.NewHandler(root, "srv@v0.0.1"))
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/root/target", nil))
		return rw
	}

	bdd.From(t).Given("pre operator returns an error response as value", func(t bdd.T) {
		t.When("request is served", func(t bdd.T) {
			rw := serve(&Deny{})
			t.Then("error response written", bdd.Equal(rw.Code, http.StatusForbidden))
			t.Then("problem media type", bdd.Equal(rw.Header().Get("Content-Type"), "application/problem+json"))
			t.Then("following operator not executed", bdd.BeFalse(executed))
		})
	})

	bdd.From(t).Given("pre operator returns nil", func(t bdd.T) {
		t.When("request is served", func(t bdd.T) {
			rw := serve(&Allow{})
			t.Then("following operator executed", bdd.BeTrue(executed))
			t.Then("204", bdd.Equal(rw.Code, http.StatusNoContent))
		})
	})
}
