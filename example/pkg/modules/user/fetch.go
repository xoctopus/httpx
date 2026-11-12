package user

import (
	"context"
	"fmt"
	"time"

	"github.com/xoctopus/httpx/pkg/httpx"
	sqlerr "github.com/xoctopus/sqlx/pkg/errors"
	"github.com/xoctopus/x/codex"
)

func GetByUserID(_ context.Context, id string) (*Info, error) {
	switch id {
	case "3":
		return nil, codex.New(sqlerr.ROLLBACK)
	default:
		u, ok := users.Load(id)
		if !ok {
			return nil, codex.Wrapf(
				ERROR__NOT_FOUND_BY_ID,
				codex.New(sqlerr.NOTFOUND), "用户ID: %s 不存在", id,
			)
		}
		return u, nil
	}
}

func GetByUsername(_ context.Context, name string) (*Info, error) {
	switch name {
	case "3":
		return nil, httpx.STATUS__INTERNAL_SERVER_ERROR.Wrap(fmt.Errorf("some error"))
	default:
		var u *Info
		users.Range(func(_ string, info *Info) bool {
			if info.Username == name {
				u = info
				return false
			}
			return true
		})
		if u == nil {
			return nil, codex.Wrapf(
				ERROR__NOT_FOUND_BY_NAME,
				codex.New(sqlerr.NOTFOUND), "用户名: %s 不存在", name)
		}
		return u, nil
	}
}

type ListParam struct {
	UserID     string    `in:"query" name:"userID,omitempty"`
	Username   string    `in:"query" name:"username,omitempty"`
	CreatedAt  time.Time `in:"query" name:"createdAt,omitempty"`
	MemberType string    `in:"query" name:"memberType"` // NOTE: required
}

type ListResult struct {
	Data  []*Info `json:"data"`
	Total int     `json:"total"`
}

func ListUsers(_ context.Context, r *ListParam) (*ListResult, error) {
	res := &ListResult{}

	users.Range(func(_ string, u *Info) bool {
		if u.UserID == r.UserID || u.Username == r.Username || u.Type == r.MemberType {
			res.Data = append(res.Data, u)
		}
		return true
	})

	res.Total = users.Len()
	return res, nil
}
