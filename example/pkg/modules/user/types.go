package user

import "github.com/xoctopus/x/contextx"

type Controller interface {
}

type tCtxUser struct{}

var (
	From  = contextx.From[tCtxUser, Controller]
	Must  = contextx.Must[tCtxUser, Controller]
	With  = contextx.With[tCtxUser, Controller]
	Carry = contextx.Carry[tCtxUser, Controller]
)

func New() Controller {
	return nil
}
