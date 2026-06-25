package middlex

import (
	"net/http"

	"github.com/xoctopus/httpx/pkg/middlex/internal/canonicalhandler"
)

// copy from https://github.com/gorilla/handlers/blob/main/canonical.go

// CanonicalHost is HTTP middleware that re-directs requests to the canonical
// domain. It accepts a domain and a status code (e.g. 301 or 302) and
// re-directs clients to this domain. The existing request path is maintained.
//
// Note: If the provided domain is considered invalid by url.Parse or otherwise
// returns an empty scheme or host, clients are not re-directed.
//
// Example:
//
//	r := mux.NewRouter()
//	canonical := handlers.CanonicalHost("http://www.gorillatoolkit.org", 302)
//	r.HandleFunc("/route", YourHandler)
//
//	log.Fatal(http.ListenAndServe(":7000", canonical(r)))
func CanonicalHost(domain string, code int) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &canonicalhandler.Canonical{
			Handler: h,
			Domain:  domain,
			Code:    code,
		}
	}
}
