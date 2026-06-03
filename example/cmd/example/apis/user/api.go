package user

import "github.com/xoctopus/httpx/pkg/httpx"

var R = httpx.NewRouter()

func init() {
	R.Register(httpx.NewRouter(&CookieRefresh{}))
	R.Register(httpx.NewRouter(&RegisterMember{}))
	R.Register(httpx.NewRouter(&DeleteMemberByID{}))
	R.Register(httpx.NewRouter(&ListMembers{}))
}
