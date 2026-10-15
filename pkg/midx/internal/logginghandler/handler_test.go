package logginghandler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/xoctopus/x/testx"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/pkg/midx/internal/logginghandler"
)

func TestOmitAuthorization(t *testing.T) {
	bdd.From(t).Given("URL with authorization query params", func(t bdd.T) {
		u, err := url.Parse("https://api.example/v1?authorization=secret&x-param-header-Authorization=token&q=1")
		testx.Expect(t, err, testx.BeNil[error]())

		t.When("OmitAuthorization is applied", func(t bdd.T) {
			got := logginghandler.OmitAuthorization(u)
			t.Then("sensitive params are stripped", bdd.Equal(got, "https://api.example/v1?q=1"))
		})
	})
}

func TestNewResponseWriter(t *testing.T) {
	bdd.From(t).Given("response writer wrapping recorder", func(t bdd.T) {
		rec := httptest.NewRecorder()
		rw := logginghandler.NewResponseWriter(rec)

		t.When("4xx status and body are written", func(t bdd.T) {
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte(" bad request "))

			t.Then("wrapper status matches", bdd.Equal(rw.StatusCode(), http.StatusBadRequest))
			t.Then("body becomes status error", bdd.ErrorEqual(rw.StatusError(), "bad request"))
			t.Then("underlying recorder status matches", bdd.Equal(rec.Code, http.StatusBadRequest))
		})
	})
}
