package user

import (
	"context"
	"strconv"
	"time"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/syncx"
)

type RegisterParam struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Info struct {
	ID           int64     `json:"id"`
	UserID       string    `json:"userID"`
	Username     string    `json:"username"`
	Password     string    `json:"password,omitempty"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"createdAt"`
	LastActiveAt time.Time `json:"lastActiveAt"`
}

// Now is the clock of the module, replaced by examples for a stable output
var Now = time.Now

var users syncx.Map[string, *Info]

func init() { Reset() }

// Reset restores the store to the initial users 0 (管理员), 1 (普通会员) and 2 (高级会员)
func Reset() {
	users = syncx.NewXmap[string, *Info]()
	for id, typ := range []string{"管理员", "普通会员", "高级会员"} {
		uid := strconv.Itoa(id)
		users.Store(uid, &Info{
			UserID:       uid,
			Username:     uid,
			Password:     uid,
			Type:         typ,
			CreatedAt:    Now(),
			LastActiveAt: Now(),
		})
	}
}

func Register(_ context.Context, r *RegisterParam) (*Info, error) {
	if users.Len() > 5 {
		return nil, codex.Errorf(ERROR__REGISTRATION_NOT_OPEN, "敬请期待")
	}

	var err error

	users.Range(func(_ string, info *Info) bool {
		if info.Username == r.Username {
			err = codex.Errorf(ERROR__USERNAME_CONFLICT, "用户名 [%s] 已存在, 请换一个试试", r.Username)
			return false
		}
		return true
	})

	if err != nil {
		return nil, err
	}

	id := int64(users.Len() + 1)
	u := &Info{
		ID:           id,
		UserID:       strconv.FormatInt(id, 10),
		Username:     r.Username,
		Password:     r.Password,
		Type:         "普通会员",
		CreatedAt:    Now(),
		LastActiveAt: Now(),
	}
	users.Store(u.UserID, u)
	u2 := *u
	u2.Password = ""

	return &u2, nil
}
