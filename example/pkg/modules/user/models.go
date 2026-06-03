package user

import (
	"context"
	"time"
)

const MaskedPassword = "--------"

// Password as a string underlying and implements SecurityStringer
type Password string

func (p Password) String() string { return string(p) }

func (p Password) SecurityString() string { return MaskedPassword }

type RegisterReq struct {
	Username string   `json:"username"`
	Password Password `json:"password"`
}

type RegisterRsp struct {
	UserID   int64
	Username string
}

/*
type Ranger[T comparable] struct {
	Gt  T `in:"query" name:"gt,omitzero"`
	Gte T `in:"query" name:"gte,omitzero"`
	Lt  T `in:"query" name:"lt,omitzero"`
	Lte T `in:"query" name:"lte,omitzero"`
}
*/

type ListReq struct {
	CreatedAt  time.Time `in:"query" name:"createdAt,omitzero"`
	UpdatedAt  time.Time `in:"query" name:"updatedAt,omitzero"`
	MemberType string    `in:"query" name:"memberType"` // required
	// TODO order and pager
}

type Data struct {
	UserID       int64     `json:"userID"`
	Username     string    `json:"username"`
	MemberType   string    `json:"memberType"`
	CreatedAt    time.Time `json:"createdAt"`
	LastActiveAt time.Time `json:"lastActiveAt"`
}

type ListRsp struct {
	Data  []Data `json:"data"`
	Total int    `json:"total"`
}

func ListMembers(_ context.Context, _ *ListReq) (*ListRsp, error) {
	return &ListRsp{
		Data: []Data{
			{1, "1", "1", time.Now(), time.Now()},
			{2, "2", "2", time.Now(), time.Now()},
		},
		Total: 2,
	}, nil
}
