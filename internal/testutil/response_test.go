package testutil_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/testutil"
)

func ExampleUnifyResponse() {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"key":"USER::NOT_FOUND"}` + "\n"))
	})
	req, _ := http.NewRequest(http.MethodGet, "/users/1", nil)

	fmt.Println(testutil.UnifyResponse(testutil.ServeHandler(h, req).Result()))

	// Output:
	// HTTP/1.1 404 Not Found
	// Content-Length: 26
	// Content-Type: application/problem+json
	//
	// {"key":"USER::NOT_FOUND"}
}

func TestBeResponse(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	Expect(t, testutil.ServeHandler(h, req).Result(), testutil.BeResponse(`
HTTP/1.1 204 No Content
`))
}

func TestUnifyResponse(t *testing.T) {
	rsp := &http.Response{
		StatusCode:    http.StatusOK,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        http.Header{},
		Body:          io.NopCloser(strings.NewReader("body")),
		ContentLength: 4,
	}
	testutil.UnifyResponse(rsp)

	Expect(t, string(must.NoErrorV(io.ReadAll(rsp.Body))), Equal("body"))
}
