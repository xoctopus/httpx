package user

import (
	"context"
	"fmt"

	"github.com/xoctopus/httpx/pkg/httpx"
)

type DeleteUserByUserID struct {
	httpx.MethodDelete `path:"/{userID}" summary:"删除用户"`

	UserID int64 `in:"path" name:"userID"`
}

func (req *DeleteUserByUserID) Output(ctx context.Context) (any, error) {
	return nil, fmt.Errorf("UnknownError")
}

// ResponseErrors implements openapi.ResponseErrorsSpecified for spec generating
func (DeleteUserByUserID) ResponseErrors() []error {
	// TODO
	return nil
}
