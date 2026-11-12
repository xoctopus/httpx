package user

import (
	"context"
	"fmt"
	"time"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/x/codex"
)

type LoginParam struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResult struct {
	UserID    string    `json:"userID"`
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func LoginValidateByPassword(ctx context.Context, r *LoginParam) (*LoginResult, error) {
	j := confjwt.MustJWT(ctx)

	u, err := GetByUsername(ctx, r.Username)
	if err != nil {
		return nil, err
	}

	if u.Password != r.Password {
		return nil, codex.New(ERROR__PASSWORD_UNMATCHED)
	}

	tok, _ := j.Generate(u.UserID)
	return &LoginResult{
		UserID:    u.UserID,
		Username:  u.Username,
		Token:     tok,
		ExpiredAt: time.Now().Add(time.Duration(j.ExpIn)),
	}, nil
}

func RefreshToken(ctx context.Context, t string) (*LoginResult, error) {
	j := confjwt.MustJWT(ctx)

	claim, err := j.Parse(t)
	if err != nil {
		return nil, codex.Wrap(ERROR__FAILED_TO_REFRESH_TOKEN, err)
	}
	id, ok := claim.Payload.(string)
	if !ok {
		return nil, codex.Wrap(ERROR__FAILED_TO_REFRESH_TOKEN, fmt.Errorf("invalid payload"))
	}

	u, err := GetByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return LoginValidateByPassword(ctx, &LoginParam{
		Username: u.Username,
		Password: u.Password,
	})
}
