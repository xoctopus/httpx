package middlex

import (
	"net/http"
	"strings"

	"github.com/xoctopus/httpx/pkg/middlex/internal/corshandler"
)

// copy from https://github.com/gorilla/handlers/blob/master/cors.go

type (
	// CORSOption represents a functional option for configuring the CORS middleware.
	CORSOption func(*corshandler.Cors) error

	OriginValidator = corshandler.OriginValidator
)

func DefaultCORS(opts ...CORSOption) func(http.Handler) http.Handler {
	return CORS(
		append([]CORSOption{
			AllowedOrigins([]string{"*"}),
			AllowedMethods([]string{
				http.MethodConnect,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
			}),
			AllowCredentials(),
			AllowedHeaders([]string{
				corshandler.CorsRequestMethodHeader,
				corshandler.CorsRequestHeadersHeader,
				"Content-Type",
				"Authorization",
				"User-Agent",
			}),
			ExposedHeaders([]string{
				"Content-Type",
				"Origin",
				"B3",
				"WWW-Authenticate",
				"Location",
				"X-Requested-With",
				"X-RateLimit-Limit", // follow https://developer.github.com/v3/rate_limit/
				"X-RateLimit-Remaining",
				"X-RateLimit-Reset",
			}),
			OptionStatusCode(http.StatusNoContent),
		}, opts...)...,
	)
}

// CORS provides Cross-Origin Resource Sharing middleware.
// Example:
//
//	import (
//	    "net/http"
//
//	    "github.com/gorilla/handlers"
//	    "github.com/gorilla/mux"
//	)
//
//	func main() {
//	    r := mux.NewRouter()
//	    r.HandleFunc("/users", UserEndpoint)
//	    r.HandleFunc("/projects", ProjectEndpoint)
//
//	    // Apply the CORS middleware to our top-level router, with the defaults.
//	    http.ListenAndServe(":8000", handlers.CORS()(r))
//	}
func CORS(opts ...CORSOption) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		ch := corshandler.DefaultCors

		for _, option := range opts {
			_ = option(&ch)
		}

		ch.SetHandler(h)
		return &ch
	}
}

//
// Functional options for configuring CORS.
//

// AllowedHeaders adds the provided headers to the list of allowed headers in a
// CORS request.
// This is an append operation so the headers Accept, Accept-Language,
// and Content-Language are always allowed.
// Content-Type must be explicitly declared if accepting Content-Types other than
// application/x-www-form-urlencoded, multipart/form-data, or text/plain.
func AllowedHeaders(headers []string) CORSOption {
	return func(ch *corshandler.Cors) error {
		for _, v := range append(corshandler.DefaultCorsHeaders, headers...) {
			normalizedHeader := http.CanonicalHeaderKey(strings.TrimSpace(v))
			if normalizedHeader == "" {
				continue
			}

			if !ch.IsMatch(normalizedHeader, ch.AllowedHeaders) {
				ch.AllowedHeaders = append(ch.AllowedHeaders, normalizedHeader)
			}
		}

		return nil
	}
}

// AllowedMethods can be used to explicitly allow methods in the
// Access-Control-Allow-Methods header.
// This is a replacement operation so you must also
// pass GET, HEAD, and POST if you wish to support those methods.
func AllowedMethods(methods []string) CORSOption {
	return func(ch *corshandler.Cors) error {
		ch.AllowedMethods = []string{}
		for _, v := range append(corshandler.DefaultCorsMethods, methods...) {
			normalizedMethod := strings.ToUpper(strings.TrimSpace(v))
			if normalizedMethod == "" {
				continue
			}

			if !ch.IsMatch(normalizedMethod, ch.AllowedMethods) {
				ch.AllowedMethods = append(ch.AllowedMethods, normalizedMethod)
			}
		}

		return nil
	}
}

// AllowedOrigins sets the allowed origins for CORS requests, as used in the
// 'Allow-Access-Control-Origin' HTTP header.
// Note: Passing in a []string{"*"} will allow any domain.
func AllowedOrigins(origins []string) CORSOption {
	return func(ch *corshandler.Cors) error {
		for _, v := range origins {
			if v == corshandler.CorsOriginMatchAll {
				ch.AllowedOrigins = []string{v}
				return nil
			}
		}

		ch.AllowedOrigins = origins
		return nil
	}
}

// AllowedOriginValidator sets a function for evaluating allowed origins in CORS requests, represented by the
// 'Allow-Access-Control-Origin' HTTP header.
func AllowedOriginValidator(fn OriginValidator) CORSOption {
	return func(ch *corshandler.Cors) error {
		ch.AllowedOriginValidator = fn
		return nil
	}
}

// OptionStatusCode sets a custom status code on the OPTIONS requests.
// Default behavior sets it to 200 to reflect best practices. This is option is not mandatory
// and can be used if you need a custom status code (i.e. 204).
//
// More information on the spec:
// https://fetch.spec.whatwg.org/#cors-preflight-fetch
func OptionStatusCode(code int) CORSOption {
	return func(ch *corshandler.Cors) error {
		ch.OptionStatusCode = code
		return nil
	}
}

// ExposedHeaders can be used to specify headers that are available
// and will not be stripped out by the user-agent.
func ExposedHeaders(headers []string) CORSOption {
	return func(ch *corshandler.Cors) error {
		ch.ExposedHeaders = []string{}
		for _, v := range headers {
			normalizedHeader := http.CanonicalHeaderKey(strings.TrimSpace(v))
			if normalizedHeader == "" {
				continue
			}

			if !ch.IsMatch(normalizedHeader, ch.ExposedHeaders) {
				ch.ExposedHeaders = append(ch.ExposedHeaders, normalizedHeader)
			}
		}

		return nil
	}
}

// MaxAge determines the maximum age (in seconds) between preflight requests. A
// maximum of 10 minutes is allowed. An age above this value will default to 10
// minutes.
func MaxAge(age int) CORSOption {
	return func(ch *corshandler.Cors) error {
		// Maximum of 10 minutes.
		if age > 600 {
			age = 600
		}

		ch.MaxAge = age
		return nil
	}
}

// IgnoreOptions causes the CORS middleware to ignore OPTIONS requests, instead
// passing them through to the next handler. This is useful when your application
// or framework has a pre-existing mechanism for responding to OPTIONS requests.
func IgnoreOptions() CORSOption {
	return func(ch *corshandler.Cors) error {
		ch.IgnoreOptions = true
		return nil
	}
}

// AllowCredentials can be used to specify that the user agent may pass
// authentication details along with the request.
func AllowCredentials() CORSOption {
	return func(ch *corshandler.Cors) error {
		ch.AllowCredentials = true
		return nil
	}
}
