package user

import (
	"context"

	"github.com/xoctopus/httpx/internal/cmd/example/modules/user"
	"github.com/xoctopus/httpx/pkg/httpx"
)

type PromoteUserLevel struct {
	httpx.MethodPut `path:"/promote/:userID" summary:"提升用户等级"`

	UserID string `in:"path" name:"userID"`
}

func (r *PromoteUserLevel) Output(ctx context.Context) (any, error) {
	return user.PromoteByUserID(ctx, r.UserID)
}
