package openapi

import (
	"github.com/xoctopus/x/contextx"
)

type tCtxOpenAPI struct{}

var (
	From  = contextx.From[tCtxOpenAPI, *OpenAPI]
	Must  = contextx.Must[tCtxOpenAPI, *OpenAPI]
	With  = contextx.With[tCtxOpenAPI, *OpenAPI]
	Carry = contextx.Carry[tCtxOpenAPI, *OpenAPI]
)
