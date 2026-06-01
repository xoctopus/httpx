package types

import (
	"net/http"

	"github.com/xoctopus/x/contextx"
)

type (
	tCtxOperationMeta struct{}
	tCtxRequest       struct{}
)

var (
	OperationMetaFrom  = contextx.From[tCtxOperationMeta, OperationMeta]
	WithOperationMeta  = contextx.With[tCtxOperationMeta, OperationMeta]
	MustOperationMeta  = contextx.Must[tCtxOperationMeta, OperationMeta]
	CarryOperationMeta = contextx.Carry[tCtxOperationMeta, OperationMeta]

	RequestFrom  = contextx.From[tCtxRequest, *http.Request]
	WithRequest  = contextx.With[tCtxRequest, *http.Request]
	MustRequest  = contextx.Must[tCtxRequest, *http.Request]
	CarryRequest = contextx.Carry[tCtxRequest, *http.Request]
)
