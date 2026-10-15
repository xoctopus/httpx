package midx_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/internal/testutil"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/midx"
)

func Example() {
	ctx := logx.With(context.Background(), logx.Discard())

	userCtx := contextx.NewT[string]()
	carriers := []contextx.Carrier{
		userCtx.Carry("alice"),
	}

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, _ := userCtx.From(r.Context())
		fmt.Fprintln(w, name)
	})

	h := httpx.ApplyMiddlewares(
		midx.InjectContext(carriers...),
		midx.LogHandler(),
		midx.DefaultCORS(),
	)(base)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil).WithContext(ctx)
	req.Header.Set("Origin", "https://app.example")
	rec := testutil.ServeHandler(h, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	fmt.Println(rec.Header().Get("Access-Control-Allow-Origin"))

	// Output:
	// 200
	// alice
	//
	// *
}
