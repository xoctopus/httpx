package openapi

import (
	"github.com/xoctopus/x/contextx"
)

type HasResponseContentType interface {
	ResponseContentType() string
}

type HasResponseStatusCode interface {
	ResponseStatusCode() int
}

type HasResponseContent interface {
	ResponseContent() any
}

type HasResponseErrors interface {
	ResponseErrors() []error
}

type tCtxOpenAPI struct{}

var (
	From  = contextx.From[tCtxOpenAPI, *OpenAPI]
	Must  = contextx.Must[tCtxOpenAPI, *OpenAPI]
	With  = contextx.With[tCtxOpenAPI, *OpenAPI]
	Carry = contextx.Carry[tCtxOpenAPI, *OpenAPI]
)
