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

func TestCanonicalHost(t *testing.T) {
	bdd.From(t).Given("request host differs from canonical domain", func(t bdd.T) {
		h := midx.CanonicalHost("https://www.example.com", http.StatusFound)(
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("next must not run")
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "http://example.com/foo?q=1", nil)
		req.Host = "example.com"

		t.When("the middleware serves the request", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("status is redirect", bdd.Equal(rec.Code, http.StatusFound))
			t.Then("location preserves path and query", bdd.Equal(rec.Header().Get("Location"), "https://www.example.com/foo?q=1"))
		})
	})

	bdd.From(t).Given("request host matches canonical domain", func(t bdd.T) {
		nextCalled := false
		h := midx.CanonicalHost("https://www.example.com", http.StatusFound)(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "https://www.example.com/foo", nil)
		req.Host = "www.example.com"

		t.When("the middleware serves the request", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("next handler runs", bdd.BeTrue(nextCalled))
			t.Then("status is OK", bdd.Equal(rec.Code, http.StatusOK))
		})
	})
}

func ExampleCanonicalHost() {
	h := midx.CanonicalHost("https://www.example.com", http.StatusFound)(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/path", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Location"))

	// Output:
	// 302
	// https://www.example.com/path
}
