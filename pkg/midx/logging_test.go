package midx_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/testutil"
	"github.com/xoctopus/httpx/pkg/midx"
)

func TestLogHandler(t *testing.T) {
	bdd.From(t).Given("logx discard logger in context", func(t bdd.T) {
		ctx := logx.With(context.Background(), logx.Discard())
		h := midx.LogHandler()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/ok", nil).WithContext(ctx)

		t.When("request is served", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("status is OK", bdd.Equal(rec.Code, http.StatusOK))
			t.Then("B3 trace header is injected", bdd.BeTrue(hasB3TraceHeader(rec)))
		})
	})
}

func TestMetricHandler(t *testing.T) {
	bdd.From(t).Given("prometheus registry", func(t bdd.T) {
		reg := prometheus.NewRegistry()
		h := midx.MetricHandler(reg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))

		t.When("metrics path is requested", func(t bdd.T) {
			rec := testutil.ServeHandler(h, httptest.NewRequest(http.MethodGet, "/.sys/metrics", nil))
			t.Then("status is OK", bdd.Equal(rec.Code, http.StatusOK))
			t.Then("content-type is text/plain", bdd.ContainsSubString(rec.Header().Get("Content-Type"), "text/plain"))
		})

		t.When("application path is requested", func(t bdd.T) {
			rec := testutil.ServeHandler(h, httptest.NewRequest(http.MethodGet, "/api", nil))
			t.Then("next handler status", bdd.Equal(rec.Code, http.StatusTeapot))
		})
	})
}

func ExampleLogHandler() {
	ctx := logx.With(context.Background(), logx.Discard())
	h := midx.LogHandler()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ok", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}

func ExampleMetricHandler() {
	reg := prometheus.NewRegistry()
	h := midx.MetricHandler(reg)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/.sys/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}

func hasB3TraceHeader(rec *httptest.ResponseRecorder) bool {
	return rec.Header().Get("X-B3-Traceid") != "" || rec.Header().Get("B3") != ""
}
