package routers

import (
	"context"
	"errors"

	"github.com/xoctopus/httpx/pkg/httpx"
)

var gForbiddenOpenAPI = false

func ForbidOpenAPI() {
	gForbiddenOpenAPI = true
}

type OpenAPI struct {
	httpx.MethodGet
}

func (o *OpenAPI) Output(ctx context.Context) (any, error) {
	if gForbiddenOpenAPI {
		// TODO should has status http.StatusForbidden
		return nil, errors.New("forbidden")
	}

	// TODO try to extract oas from context

	return nil, errors.New("forbidden")
}

func (o *OpenAPI) ResponseContentType() string {
	return "application/json"
}
