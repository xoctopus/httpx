package pprofhandler

import (
	"net/http"
	"net/http/pprof"
	"strings"
)

type Handler struct {
	Enabled bool
	Next    http.Handler
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.Enabled && strings.HasPrefix(req.URL.Path, "/.sys/debug/pprof") {
		switch req.URL.Path {
		case "/.sys/debug/pprof/cmdline":
			pprof.Cmdline(rw, req)
			return
		case "/.sys/debug/pprof/profile":
			pprof.Profile(rw, req)
			return
		case "/.sys/debug/pprof/symbol":
			pprof.Symbol(rw, req)
			return
		case "/.sys/debug/pprof/trace":
			pprof.Trace(rw, req)
			return
		default:
			// trim /.sys for make pprof.Index work
			req.URL.Path = req.URL.Path[len("/.sys"):]
			pprof.Index(rw, req)
			return
		}
	}
	h.Next.ServeHTTP(rw, req)
}
