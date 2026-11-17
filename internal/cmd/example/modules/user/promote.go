package user

import "context"

type PromotionResult struct {
	UserID string `json:"userID"`
	Prev   string `json:"prev"`
	Curr   string `json:"curr"`
}

func PromoteByUserID(ctx context.Context, id string) (any, error) {
	u, err := GetByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	r := &PromotionResult{
		UserID: u.UserID,
		Prev:   u.Type,
		Curr:   "",
	}
	u.Type = "高级会员"
	r.Curr = u.Type
	return r, nil
}
