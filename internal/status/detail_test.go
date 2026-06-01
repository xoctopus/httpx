package status_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/pkg/httpx"
)

func ExampleAsDescription() {
	// error
	e := status.WrapCode(errors.New("x"), http.StatusInternalServerError)
	fmt.Println(e)

	// wrapped error
	e = status.WrapCode(errors.New("e1"), http.StatusForbidden)
	de := status.AsDescription(e, "server-name", "query")
	fmt.Println(de)

	// user error
	e = errors.New("e2")
	de = status.AsDescription(e, "server-name", "query")
	fmt.Println(de)

	// validation error
	x := errors.New("e3")

	e = validation.WrapLocationError(x, "query")
	de = status.AsDescription(e, "server-name", "")
	OutputDescription(de)

	e = validation.WrapPositionError(x, "field")
	de2 := status.AsDescription(e, "server-name", "")
	OutputDescription(de2)

	de3 := status.AsDescription(de2, "", "")
	de3.Location = "body"
	OutputDescription(de3)

	// Output:
	// INTERNAL_SERVER_ERROR{message="x",status=500}
	// FORBIDDEN{message="e1"}
	// INTERNAL_SERVER_ERROR{message="e2"}
	// BAD_REQUEST{message="e3"}
	//   in:   query
	//   code: 400
	//   text: BAD_REQUEST
	// BAD_REQUEST{message="e3"}
	//   pos:  field
	//   code: 400
	//   text: BAD_REQUEST
	// BAD_REQUEST{message="e3"}
	//   in:   body
	//   pos:  field
	//   code: 400
	//   text: BAD_REQUEST
}

func ExampleUnmarshalResponse() {
	// unmarshal from response body
	rspraw, _ := json.Marshal(status.Response{
		Code: http.StatusUnauthorized,
		Errors: []*status.Description{
			{
				Text:     "BAD_REQUEST",
				Message:  "",
				Location: "body",
				Position: "username",
				Source:   "srv-auth-gateway@v0.0.1",
				Status:   httpx.STATUS__BAD_REQUEST,
			},
		},
		Extra: map[string]any{
			"title":  "用户名不合法",
			"detail": "请输入8-16位, 以应为字母开头且仅包含英文字符和数字",
		},
	})
	de := status.UnmarshalResponse(http.StatusUnauthorized, rspraw)
	OutputDescription(de)

	rspraw, _ = json.Marshal(status.Response{
		Code:    http.StatusBadRequest,
		Message: "用户名不合法\n请输入8-16位, 以应为字母开头且仅包含英文字符和数字",
	})
	de = status.UnmarshalResponse(http.StatusBadRequest, rspraw)
	OutputDescription(de)

	rspraw, _ = json.Marshal(status.Response{Code: http.StatusBadRequest})
	de = status.UnmarshalResponse(http.StatusBadRequest, rspraw)
	OutputDescription(de)

	de = status.UnmarshalResponse(http.StatusBadRequest, []byte("用户名不合法"))
	OutputDescription(de)

	// Output:
	// BAD_REQUEST{message="用户名不合法\n请输入8-16位, 以应为字母开头且仅包含英文字符和数字"}
	//   in:   body
	//   pos:  username
	//   code: 401
	//   text: UNAUTHORIZED
	// BAD_REQUEST{message="用户名不合法\n请输入8-16位, 以应为字母开头且仅包含英文字符和数字"}
	//   code: 400
	//   text: BAD_REQUEST
	// BAD_REQUEST{message="BAD_REQUEST"}
	//   code: 400
	//   text: BAD_REQUEST
	// INTERNAL_SERVER_ERROR{message="用户名不合法"}
	//   code: 500
	//   text: INTERNAL_SERVER_ERROR
}

func OutputDescription(de *status.Description) {
	fmt.Println(de)
	if len(de.Location) > 0 {
		fmt.Println("  in:  ", de.Location)
	}
	if len(de.Position) > 0 {
		fmt.Println("  pos: ", de.Position)
	}
	fmt.Println("  code:", de.StatusCode())
	fmt.Println("  text:", de.StatusText())
}
