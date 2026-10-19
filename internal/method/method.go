// Package method defines optional capabilities for HTTP verbs on operator types.
//
// Route and client operators embed or implement these interfaces so [route.Meta],
// OpenAPI generation, and [transport.Outgoing] can read the verb without
// reflection. Values should match net/http method constants (for example
// http.MethodGet). The public alias is [httpx.MethodDescriptor]; common
// implementations are [httpx.MethodGet], [httpx.MethodPost], and siblings.
package method

// Describer reports the HTTP method for an operator or outgoing request type.
type Describer interface {
	Method() string
}
