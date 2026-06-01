package user

import "github.com/xoctopus/httpx/pkg/httpx"

var R = httpx.NewRouter()

func init() {
	R.Register(httpx.NewRouter())
}
