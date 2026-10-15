package midx_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/testutil"
	"github.com/xoctopus/httpx/pkg/midx"
)

func TestDefaultCORS(t *testing.T) {
	bdd.From(t).Given("DefaultCORS with a GET and Origin header", func(t bdd.T) {
		nextCalled := false
		h := midx.DefaultCORS()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api", nil)
		req.Header.Set("Origin", "https://app.example")

		t.When("the request is served", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("next handler runs", bdd.BeTrue(nextCalled))
			t.Then("allow origin is wildcard", bdd.Equal(rec.Header().Get("Access-Control-Allow-Origin"), "*"))
			t.Then("credentials allowed", bdd.Equal(rec.Header().Get("Access-Control-Allow-Credentials"), "true"))
		})
	})

	bdd.From(t).Given("DefaultCORS with OPTIONS preflight", func(t bdd.T) {
		h := midx.DefaultCORS()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("next must not run on preflight")
		}))

		req := httptest.NewRequest(http.MethodOptions, "/api", nil)
		req.Header.Set("Origin", "https://app.example")
		req.Header.Set("Access-Control-Request-Method", http.MethodPut)

		t.When("preflight is served", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("status is 204", bdd.Equal(rec.Code, http.StatusNoContent))
			t.Then("allow origin is wildcard", bdd.Equal(rec.Header().Get("Access-Control-Allow-Origin"), "*"))
		})
	})
}

func ExampleDefaultCORS() {
	h := midx.DefaultCORS()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Origin", "https://app.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Access-Control-Allow-Origin"))

	// Output:
	// 200
	// *
}
