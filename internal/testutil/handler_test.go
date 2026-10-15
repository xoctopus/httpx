package testutil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/testutil"
)

func TestServeHandler(t *testing.T) {
	bdd.From(t).Given("a handler that writes teapot", func(t bdd.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)

		t.When("ServeHandler runs the chain", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("recorder captures status", bdd.Equal(rec.Code, http.StatusTeapot))
		})
	})
}
