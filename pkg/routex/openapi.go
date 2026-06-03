package routex

import (
	"context"
	"fmt"

	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/types"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/spec/openapi"
)

var gForbiddenOpenAPI = false

func ForbidOpenAPI() {
	gForbiddenOpenAPI = true
}

var (
	ErrOpenAPIForbidden = status.Wrap(fmt.Errorf("openapi is forbidden"), httpx.STATUS__FORBIDDEN)
	ErrOpenAPINotFound  = status.Wrap(fmt.Errorf("openapi view not found"), httpx.STATUS__NOTFOUND)
)

type OpenAPIProvider interface {
	OpenAPI() *openapi.OpenAPI
}

type OpenAPI struct {
	httpx.MethodGet
}

func (o *OpenAPI) Output(ctx context.Context) (any, error) {
	if gForbiddenOpenAPI {
		return nil, ErrOpenAPIForbidden
	}

	if x, ok := types.OperationMetaProviderFrom(ctx); ok {
		if u, ok := x.(OpenAPIProvider); ok {
			return &openapi.Payload{OpenAPI: *u.OpenAPI()}, nil
		}
	}

	return nil, ErrOpenAPIForbidden
}

func (o *OpenAPI) ResponseContentType() string {
	return "application/json"
}

// TODO should ServeFS for openapi view
var openapiView httpx.Upgrader

func SetOpenAPIViewContents(u httpx.Upgrader) {
	openapiView = u
}

type OpenAPIView struct {
	httpx.MethodGet `path:"/_view/{href...}"`
}

func (o *OpenAPIView) Output(ctx context.Context) (any, error) {
	if openapiView == nil {
		return nil, ErrOpenAPINotFound
	}
	return openapiView, nil
}
