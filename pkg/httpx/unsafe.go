package httpx

import (
	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/internal/types"
)

// NOTE:
// These exported context methods have the capability to hijack and retrieve all
// data from the original HTTP request within the request pipeline.
// Please ensure you fully understand the implications before using these methods.

var (
	OperationMetaFrom  = types.OperationMetaFrom
	WithOperationMeta  = types.WithOperationMeta
	MustOperationMeta  = types.MustOperationMeta
	CarryOperationMeta = types.CarryOperationMeta

	RequestFrom  = types.RequestFrom
	WithRequest  = types.WithRequest
	MustRequest  = types.MustRequest
	CarryRequest = types.CarryRequest

	OperationMetaProviderFrom  = types.OperationMetaProviderFrom
	WithOperationMetaProvider  = types.WithOperationMetaProvider
	MustOperationMetaProvider  = types.MustOperationMetaProvider
	CarryOperationMetaProvider = types.CarryOperationMetaProvider
)

type tCtxServers struct{}

var (
	ServersFrom  = contextx.From[tCtxServers, []string]
	WithServers  = contextx.With[tCtxServers, []string]
	MustServers  = contextx.Must[tCtxServers, []string]
	CarryServers = contextx.Carry[tCtxServers, []string]
)
