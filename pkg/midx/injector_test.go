package midx_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/testutil"
	"github.com/xoctopus/httpx/pkg/midx"
)

func TestInjectContext(t *testing.T) {
	bdd.From(t).Given("a carrier value 42", func(t bdd.T) {
		c := contextx.NewT[int]()
		var got int
		h := midx.InjectContext(c.Carry(42))(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			got, _ = c.From(r.Context())
		}))

		t.When("the middleware serves a request", func(t bdd.T) {
			testutil.ServeHandler(h, httptest.NewRequest(http.MethodGet, "/", nil))
			t.Then("the handler reads the injected value", bdd.Equal(got, 42))
		})
	})
}

func ExampleInjectContext() {
	c := contextx.NewT[int]()
	h := midx.InjectContext(c.Carry(7))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, _ := c.From(r.Context())
		fmt.Println(v)
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	// Output:
	// 7
}
