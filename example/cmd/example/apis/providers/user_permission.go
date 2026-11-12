package providers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xoctopus/httpx/pkg/httpx"
)

// UserPermission validation for user permission
// This is a simple middleware implementation, only UserID = "0" has permission.
type UserPermission struct{}

func (*UserPermission) Output(ctx context.Context) (any, error) {
	u, ok := MayCurrentUserInfoFrom(ctx)
	if !ok || u.UserID != "0" {
		return httpx.ErrorResponse(
			fmt.Errorf("PermissionDenied"),
			httpx.WithResponseStatusCode(http.StatusForbidden),
			httpx.WithResponseTitle("无权操作"),
			httpx.WithResponseDetail("需要管理员权限操作[%s]", ""), // 最后一个operator summary
		), nil
	}
	return nil, nil
}
