package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/httpx/example/pkg/modules/user"
	"github.com/xoctopus/httpx/pkg/httpx"
)

type ListMembers struct {
	httpx.MethodGet `path:""`

	user.ListReq
}

func (req *ListMembers) Output(ctx context.Context) (any, error) {
	fmt.Println(string(must.NoErrorV(json.Marshal(req.ListReq))))
	return user.ListMembers(ctx, &req.ListReq)
}

// ResponseContent implements openapi.ResponseContentSpecified for spec generating
func (req *ListMembers) ResponseContent() any {
	return new(user.ListRsp)
}
