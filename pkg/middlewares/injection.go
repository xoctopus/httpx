package middlewares

import (
	"net/http"

	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/pkg/httpx"
)

func InjectContext(carriers ...contextx.Carrier) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := contextx.Compose(carriers...)(r.Context())
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}
