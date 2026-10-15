package testutil

import (
	"net/http"
	"net/http/httptest"
)

// ServeHandler executes h against req with [httptest.NewRecorder] and returns the recorder.
func ServeHandler(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// NewHTTPServer is [httptest.NewServer] for handlers that need a real listener (Client.Do, TLS, etc.).
func NewHTTPServer(h http.Handler) *httptest.Server {
	return httptest.NewServer(h)
}
