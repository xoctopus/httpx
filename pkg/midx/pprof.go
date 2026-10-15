package midx

import (
	"net/http"

	"github.com/xoctopus/httpx/pkg/midx/internal/pprofhandler"
)

func PProfHandler(enabled bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &pprofhandler.Handler{
			Enabled: enabled,
			Next:    handler,
		}
	}
}
