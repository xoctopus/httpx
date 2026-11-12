package user

import (
	"context"

	"github.com/xoctopus/httpx/example/pkg/modules/user"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/x/slicex"
)

type ListUsers struct {
	httpx.MethodGet `path:"" summary:"用户列表"`

	user.ListParam
}

func (req *ListUsers) Output(ctx context.Context) (any, error) {
	res, err := user.ListUsers(ctx, &req.ListParam)
	if err != nil {
		return nil, err
	}
	data := slicex.Mapping(res.Data, func(info *user.Info) *user.Info {
		u2 := *info
		u2.Password = ""
		return &u2
	})
	res.Data = data
	return res, nil
}

// ResponseContent implements openapi.ResponseContentSpecified for spec generating
func (req *ListUsers) ResponseContent() any {
	return new(user.ListResult)
}

type FetchByUsername struct {
	httpx.MethodGet `path:"/name/:name" summary:"按用户名查询用户信息"`

	Name string `in:"path" name:"name"`
}

func (r *FetchByUsername) ResponseContent() any {
	return new(user.Info)
}

func (r *FetchByUsername) Output(ctx context.Context) (any, error) {
	u, err := user.GetByUsername(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	u2 := *u
	u2.Password = ""
	return &u2, nil
}
