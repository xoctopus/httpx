package pprofhandler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/testutil"
	"github.com/xoctopus/httpx/pkg/midx/internal/pprofhandler"
)

func TestHandler(t *testing.T) {
	bdd.From(t).Given("handler enabled with next teapot handler", func(t bdd.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
		h := &pprofhandler.Handler{Enabled: true, Next: next}

		t.When("cmdline path is requested", func(t bdd.T) {
			req := httptest.NewRequest(http.MethodGet, "/.sys/debug/pprof/cmdline", nil)
			rec := testutil.ServeHandler(h, req)
			t.Then("pprof responds OK", bdd.Equal(rec.Code, http.StatusOK))
		})
	})
}
