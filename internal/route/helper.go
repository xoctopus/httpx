package route

import (
	"net/http"
	"strings"
)

func MethodOverwrite() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if method := req.Header.Get("x-http-method-override"); method != "" {
				req.Method = strings.ToUpper(method)
			}
			next.ServeHTTP(rw, req)
		})
	}
}
