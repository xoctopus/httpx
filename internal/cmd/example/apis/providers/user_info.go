package providers

import (
	"context"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/httpx/internal/cmd/example/modules/user"
)

type tCtxCurrentUserInfo struct{}

// CurrentUserInfo this a simple middleware for query user info model from token.
type CurrentUserInfo struct {
}

func (*CurrentUserInfo) ContextKey() any {
	return tCtxCurrentUserInfo{}
}

func (r *CurrentUserInfo) Output(ctx context.Context) (any, error) {
	id := confjwt.AuthorizationFrom(ctx)
	return user.GetByUserID(ctx, id)
}

func CurrentUserInfoFrom(ctx context.Context) *user.Info {
	return must.BeTrueV(MayCurrentUserInfoFrom(ctx))
}

func MayCurrentUserInfoFrom(ctx context.Context) (*user.Info, bool) {
	x, ok := ctx.Value(tCtxCurrentUserInfo{}).(*user.Info)
	return x, ok
}
