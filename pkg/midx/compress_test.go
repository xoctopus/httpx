package midx_test

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xoctopus/x/testx"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/testutil"
	"github.com/xoctopus/httpx/pkg/midx"
)

func TestDefaultCompress(t *testing.T) {
	bdd.From(t).Given("client accepts gzip", func(t bdd.T) {
		body := []byte("compress me compress me compress me")
		h := midx.DefaultCompress()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(body)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		t.When("the response is served", func(t bdd.T) {
			rec := testutil.ServeHandler(h, req)
			t.Then("content-encoding is gzip", bdd.Equal(rec.Header().Get("Content-Encoding"), "gzip"))
			t.Then("vary includes Accept-Encoding", bdd.Equal(rec.Header().Get("Vary"), "Accept-Encoding"))

			r, err := gzip.NewReader(rec.Body)
			testx.Expect(t, err, testx.BeNil[error]())
			defer func() { _ = r.Close() }()
			out, err := io.ReadAll(r)
			testx.Expect(t, err, testx.BeNil[error]())
			t.Then("body round-trips", bdd.Equal(out, body))
		})
	})
}

func ExampleDefaultCompress() {
	body := []byte("hello world")
	h := midx.DefaultCompress()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("Content-Encoding"))

	// Output:
	// gzip
}
