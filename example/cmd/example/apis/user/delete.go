package user

import (
	"context"

	"github.com/xoctopus/httpx/example/pkg/modules/user"
	"github.com/xoctopus/httpx/pkg/httpx"
)

type DeleteMemberByID struct {
	httpx.MethodDelete `path:"/{userID}"`

	UserID int64 `in:"path" name:"userID"`
}

func (req *DeleteMemberByID) Output(ctx context.Context) (any, error) {
	return httpx.ErrorResponse(httpx.STATUS__FORBIDDEN.Wrap(&user.ErrorForbidden{})), nil
}

// ResponseErrors implements openapi.ResponseErrorsSpecified for spec generating
func (DeleteMemberByID) ResponseErrors() []error {
	return []error{
		user.ErrorForbidden{},
		user.ErrorNotFound{},
	}
}
