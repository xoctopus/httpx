package request

import "github.com/xoctopus/x/contextx"

type tCtxRequestInfo struct{}

var (
	Extract = contextx.From[tCtxRequestInfo, Request]
	With    = contextx.With[tCtxRequestInfo, Request]
	Must    = contextx.Must[tCtxRequestInfo, Request]
	Carry   = contextx.Carry[tCtxRequestInfo, Request]
)
