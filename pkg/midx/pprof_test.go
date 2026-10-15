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

func TestPProfHandler(t *testing.T) {
	bdd.From(t).Given("PProfHandler disabled", func(t bdd.T) {
		nextCalled := false
		h := midx.PProfHandler(false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusTeapot)
		}))

		req := httptest.NewRequest(http.MethodGet, "/.sys/debug/pprof/", nil)

		t.When("pprof path is requested", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("next handler runs", bdd.BeTrue(nextCalled))
			t.Then("status from next", bdd.Equal(rec.Code, http.StatusTeapot))
		})
	})

	bdd.From(t).Given("PProfHandler enabled", func(t bdd.T) {
		nextCalled := false
		h := midx.PProfHandler(true)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			nextCalled = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/.sys/debug/pprof/", nil)

		t.When("pprof index is requested", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("next handler does not run", bdd.BeFalse(nextCalled))
			t.Then("pprof index responds OK", bdd.Equal(rec.Code, http.StatusOK))
		})
	})
}

func ExamplePProfHandler() {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := midx.PProfHandler(true)(next)

	req := httptest.NewRequest(http.MethodGet, "/.sys/debug/pprof/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}
