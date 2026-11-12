package status_test

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/validation"
)

func ExampleAsDetail() {
	// domain code wraps a cause: code, key, title and detail come from the domain code
	OutputDetail(status.AsDetail(codex.Wrapf(USER__NOT_FOUND, errors.New("sql: no rows"), "user %d not found", 9), "srv"))

	// plain error with an explicit status: no identity layer
	OutputDetail(status.AsDetail(status.WrapCode(errors.New("some error"), http.StatusServiceUnavailable), "srv"))

	// plain error: 500
	OutputDetail(status.AsDetail(errors.New("UnknownError"), "srv"))

	// request located error: 400
	OutputDetail(status.AsDetail(validation.PrefixPointer(validation.WrapPointer(errors.New("e3"), "/items/0/age"), "/body"), "srv"))

	// Output:
	// USER::NOT_FOUND{user 9 not found}
	//   title:   用户不存在
	//   code:    404
	// SERVICE_UNAVAILABLE{SERVICE_UNAVAILABLE{message="some error",status=503}}
	//   code:    503
	// INTERNAL_SERVER_ERROR{UnknownError}
	//   code:    500
	// BAD_REQUEST{e3: [pointer:/body/items/0/age]}
	//   pointer: /body/items/0/age
	//   code:    400
}

func ExampleUnmarshalResponse() {
	// RFC 9457 style problem details, sub problems are kept in errors
	rspraw, _ := json.Marshal(status.Response{
		Key:    "AUTH__INVALID_USERNAME",
		Code:   1,
		Title:  "用户名不合法",
		Detail: "请输入8-16位, 以应为字母开头且仅包含英文字符和数字",
		Errors: []*status.Detail{{
			Key:     "BAD_REQUEST",
			Pointer: "/body/username",
			Src:     "srv-auth-gateway@v0.0.1",
			Status:  http.StatusBadRequest,
		}},
	})
	de := status.UnmarshalResponse(http.StatusBadRequest, rspraw)
	OutputDetail(de)
	OutputDetail(de.Errors[0])

	// compatible with {code, msg} and {message}
	de = status.UnmarshalResponse(http.StatusUnauthorized, []byte(`{"code":2,"msg":"未登录"}`))
	OutputDetail(de)
	de = status.UnmarshalResponse(http.StatusUnauthorized, []byte(`{"message":"未登录"}`))
	OutputDetail(de)

	// empty body
	de = status.UnmarshalResponse(http.StatusNotFound, []byte(`{}`))
	OutputDetail(de)
	de = status.UnmarshalResponse(http.StatusNotFound, nil)
	OutputDetail(de)

	// non JSON body
	de = status.UnmarshalResponse(http.StatusBadRequest, []byte("用户名不合法"))
	OutputDetail(de)

	// Output:
	// AUTH__INVALID_USERNAME{请输入8-16位, 以应为字母开头且仅包含英文字符和数字}
	//   title:   用户名不合法
	//   code:    400
	// BAD_REQUEST{}
	//   pointer: /body/username
	//   code:    0
	// UNAUTHORIZED{未登录}
	//   title:   未登录
	//   code:    401
	// UNAUTHORIZED{未登录}
	//   title:   未登录
	//   code:    401
	// NOT_FOUND{Not Found}
	//   title:   Not Found
	//   code:    404
	// NOT_FOUND{Not Found}
	//   title:   Not Found
	//   code:    404
	// INTERNAL_SERVER_ERROR{用户名不合法}
	//   title:   Internal Server Error
	//   code:    500
}

func OutputDetail(de *status.Detail) {
	fmt.Println(de)
	if len(de.Title) > 0 {
		fmt.Println("  title:  ", de.Title)
	}
	if len(de.Pointer) > 0 {
		fmt.Println("  pointer:", de.Pointer)
	}
	fmt.Println("  code:   ", de.StatusCode())
}
