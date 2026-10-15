package corshandler

import (
	"net/http"
	"strconv"
	"strings"
)

// OriginValidator takes an origin string and returns whether or not that origin is allowed.
type OriginValidator func(string) bool

const (
	CorsOptionMethod           string = "OPTIONS"
	CorsAllowOriginHeader      string = "Access-Control-Allow-Origin"
	CorsExposeHeadersHeader    string = "Access-Control-Expose-Headers"
	CorsMaxAgeHeader           string = "Access-Control-Max-Age"
	CorsAllowMethodsHeader     string = "Access-Control-Allow-Methods"
	CorsAllowHeadersHeader     string = "Access-Control-Allow-Headers"
	CorsAllowCredentialsHeader string = "Access-Control-Allow-Credentials"
	CorsRequestMethodHeader    string = "Access-Control-Request-Method"
	CorsRequestHeadersHeader   string = "Access-Control-Request-Headers"
	CorsOriginHeader           string = "Origin"
	CorsVaryHeader             string = "Vary"
	CorsOriginMatchAll         string = "*"
)

var (
	DefaultCorsOptionStatusCode = 200
	DefaultCorsMethods          = []string{"GET", "HEAD", "POST"}
	DefaultCorsHeaders          = []string{"Accept", "Accept-Language", "Content-Language", "Origin"}

	DefaultCors = Cors{
		AllowedMethods:   DefaultCorsMethods,
		AllowedHeaders:   DefaultCorsHeaders,
		AllowedOrigins:   []string{},
		OptionStatusCode: DefaultCorsOptionStatusCode,
	}
)

type Cors struct {
	h                      http.Handler
	AllowedHeaders         []string
	AllowedMethods         []string
	AllowedOrigins         []string
	AllowedOriginValidator OriginValidator
	ExposedHeaders         []string
	MaxAge                 int
	IgnoreOptions          bool
	AllowCredentials       bool
	OptionStatusCode       int
}

func (ch *Cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get(CorsOriginHeader)
	if !ch.IsOriginAllowed(origin) {
		if r.Method != CorsOptionMethod || ch.IgnoreOptions {
			ch.h.ServeHTTP(w, r)
		}
		return
	}

	if r.Method == CorsOptionMethod {
		if ch.IgnoreOptions {
			ch.h.ServeHTTP(w, r)
			return
		}

		if _, ok := r.Header[CorsRequestMethodHeader]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		method := r.Header.Get(CorsRequestMethodHeader)
		if !ch.IsMatch(method, ch.AllowedMethods) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		requestHeaders := strings.Split(r.Header.Get(CorsRequestHeadersHeader), ",")
		allowedHeaders := make([]string, 0)
		for _, v := range requestHeaders {
			canonicalHeader := http.CanonicalHeaderKey(strings.TrimSpace(v))
			if canonicalHeader == "" {
				continue
			}

			if !ch.IsMatch(canonicalHeader, ch.AllowedHeaders) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			allowedHeaders = append(allowedHeaders, canonicalHeader)
		}

		if len(allowedHeaders) > 0 {
			w.Header().Set(CorsAllowHeadersHeader, strings.Join(allowedHeaders, ","))
		}

		if ch.MaxAge > 0 {
			w.Header().Set(CorsMaxAgeHeader, strconv.Itoa(ch.MaxAge))
		}

		if !ch.IsMatch(method, DefaultCorsMethods) {
			w.Header().Set(CorsAllowMethodsHeader, method)
		}
	} else {
		if len(ch.ExposedHeaders) > 0 {
			w.Header().Set(CorsExposeHeadersHeader, strings.Join(ch.ExposedHeaders, ","))
		}
	}

	if ch.AllowCredentials {
		w.Header().Set(CorsAllowCredentialsHeader, "true")
	}

	if len(ch.AllowedOrigins) > 1 {
		w.Header().Set(CorsVaryHeader, CorsOriginHeader)
	}

	returnOrigin := origin
	if ch.AllowedOriginValidator == nil && len(ch.AllowedOrigins) == 0 {
		returnOrigin = "*"
	} else {
		for _, o := range ch.AllowedOrigins {
			// A configuration of * is different than explicitly setting an allowed
			// origin. Returning arbitrary origin headers in an access control allow
			// origin header is unsafe and is not required by any use case.
			if o == CorsOriginMatchAll {
				returnOrigin = "*"
				break
			}
		}
	}

	w.Header().Set(CorsAllowOriginHeader, returnOrigin)

	if r.Method == CorsOptionMethod {
		w.WriteHeader(ch.OptionStatusCode)
		return
	}
	ch.h.ServeHTTP(w, r)
}

func (ch *Cors) IsOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	if ch.AllowedOriginValidator != nil {
		return ch.AllowedOriginValidator(origin)
	}

	if len(ch.AllowedOrigins) == 0 {
		return true
	}

	for _, allowedOrigin := range ch.AllowedOrigins {
		if allowedOrigin == origin || allowedOrigin == CorsOriginMatchAll {
			return true
		}
	}

	return false
}

func (ch *Cors) IsMatch(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}

func (ch *Cors) SetHandler(h http.Handler) {
	ch.h = h
}
