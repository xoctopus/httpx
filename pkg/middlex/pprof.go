package middlex

import (
	"net/http"

	"github.com/xoctopus/httpx/pkg/middlex/internal/pprofhandler"
)

func PProfHandler(enabled bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &pprofhandler.Handler{
			Enabled: enabled,
			Next:    handler,
		}
	}
}
