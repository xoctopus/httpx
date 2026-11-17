package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/logx"
	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/httpx/internal/cmd/example/apis"
	"github.com/xoctopus/httpx/internal/cmd/example/apis/login"
	_ "github.com/xoctopus/httpx/internal/cmd/example/apis/providers" // named by examples
	"github.com/xoctopus/httpx/internal/cmd/example/apis/user"
	muser "github.com/xoctopus/httpx/internal/cmd/example/modules/user"
	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/testutil"
	"github.com/xoctopus/httpx/pkg/confhttp"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/midx"
	"github.com/xoctopus/httpx/pkg/routex"
)

// Acceptance tests of the response-model spec (openspec/changes/define-response-model,
// openspec/changes/split-error-location), driven through the example service as
// Examples: each one sends a request by confhttp.Client, decodes the response by
// Result.Into as a downstream service would do, and prints the raw response
// (status line, headers, body) as the expected output.
//
// The output is stable: the module clock is fixed (muser.Now), the Date header
// is removed and tokens are masked as <jwt>. Examples depending on the user
// store call muser.Reset first.

var (
	jwt = &confjwt.JWT{Issuer: "example", ExpIn: types.Duration(time.Hour), SignKey: "x"}

	endpoint string           // http://127.0.0.1:{port}/example/v0
	root     *confhttp.Client // {endpoint}
	users    *confhttp.Client // {endpoint}/user

	// ctx for requests not shown; Client.Do requires an http client in context
	ctx = confhttp.WithHttpClient(context.Background(), &http.Client{})
)

func TestMain(m *testing.M) {
	ctx := logx.With(ctx, logx.NewStd())

	muser.Now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	muser.Reset()

	// the same assembly as main()
	h := must.NoErrorV(routex.NewHandler(apis.Root, "example@v0.0.1", midx.DefaultLogHandler()))
	h = httpx.ApplyMiddlewares(
		midx.InjectContext([]contextx.Carrier{confjwt.CarryJWT(jwt)}...),
		midx.DefaultCORS(),
	)(h)

	l := must.NoErrorV(net.Listen("tcp", "127.0.0.1:0"))
	addr := l.Addr().String()
	must.NoError(l.Close())
	go func() { _ = confhttp.ListenAndServe(ctx, addr, h) }()

	endpoint = "http://" + addr + "/example/v0"
	root = &confhttp.Client{Endpoint: endpoint}
	users = &confhttp.Client{Endpoint: endpoint + "/user"}
	must.NoError(root.Init(ctx))
	must.NoError(users.Init(ctx))

	// make sure the service is up
	for i := 0; ; i++ {
		if _, err := root.Do(ctx, &apis.HealthCheck{}).Into(nil); err == nil {
			break
		}
		must.BeTrue(i < 100)
		time.Sleep(20 * time.Millisecond)
	}

	os.Exit(m.Run())
}

// keeper keeps the raw response while Result.Into still reads the body
type keeper struct{ dump string }

func (k *keeper) RoundTrip(req *http.Request) (*http.Response, error) {
	rsp, err := http.DefaultTransport.RoundTrip(req)
	if err == nil {
		k.dump = testutil.UnifyResponse(rsp) // DumpResponse restores the body for Into
	}
	return rsp, err
}

var (
	reDate = regexp.MustCompile(`(?m)^Date: .*\n`)
	reJWT  = regexp.MustCompile(`eyJ[\w-]+\.[\w-]+\.[\w-]+`)
)

// show sends req (an operator value or *http.Request), decodes the response into
// out and prints the raw response; it returns the error of Result.Into
func show(c *confhttp.Client, req, out any, metas ...httpx.Metadata) error {
	k := &keeper{}
	ctx := confhttp.WithHttpClient(context.Background(), &http.Client{Transport: k})
	_, err := c.Do(ctx, req, metas...).Into(out)

	head, body, _ := strings.Cut(reJWT.ReplaceAllString(reDate.ReplaceAllString(k.dump, ""), "<jwt>"), "\n\n")
	fmt.Println(head)
	if body != "" {
		buf := &bytes.Buffer{}
		if json.Indent(buf, []byte(body), "", "  ") == nil {
			body = buf.String()
		}
		fmt.Println()
		fmt.Println(body)
	}
	return err
}

func raw(method, path, contentType, body string) *http.Request {
	req := must.NoErrorV(http.NewRequest(method, endpoint+path, strings.NewReader(body)))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req
}

func token(id string) string { return must.NoErrorV(jwt.Generate(id)) }

func bearer(id string) httpx.Metadata {
	return httpx.Metadata{"Authorization": {"Bearer " + token(id)}}
}

func loginBy(username, password string) *login.LoginByUsername {
	r := &login.LoginByUsername{}
	r.Username, r.Password = username, password
	return r
}

func listUsers(memberType string) *user.ListUsers {
	r := &user.ListUsers{}
	r.MemberType = memberType
	return r
}

func register(username string) *login.UserRegister {
	return &login.UserRegister{Body: muser.RegisterParam{Username: username, Password: "x"}}
}

// spec: 成功响应的状态码 / 返回 nil
func ExampleHealthCheck() {
	_ = show(root, &apis.HealthCheck{}, nil)

	// Output:
	// HTTP/1.1 204 No Content
	// Server: example@v0.0.1(HealthCheck)
}

// spec: 成功响应的状态码 / 其他方法返回数据; Content-Type
func ExampleLoginByUsername() {
	_ = show(root, loginBy("1", "1"), &muser.LoginResult{})

	// Output:
	// HTTP/1.1 200 OK
	// Content-Length: 207
	// Content-Type: application/json; charset=utf-8
	// Server: example@v0.0.1(LoginByUsername)
	//
	// {
	//   "userID": "1",
	//   "username": "1",
	//   "token": "<jwt>",
	//   "expiredAt": "2026-01-01T01:00:00Z"
	// }
}

// spec: 标识层 / 领域错误包装组件错误; 本次出错的说明 / 带说明的领域错误; 内部信息不外泄
// (SQLERROR / NOTFOUND not in body). the decoded *status.Detail keeps the status
// code and the fields of the body.
func ExampleLoginByUsername_sceneUserNotFoundByName() {
	err := show(root, loginBy("nobody", "x"), &muser.LoginResult{})
	d, _ := errors.AsType[*status.Detail](err)
	fmt.Println()
	fmt.Println(d.StatusCode(), d.Code, d.Key, d.Title, d.Detail)

	// Output:
	// HTTP/1.1 404 Not Found
	// Content-Length: 255
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(LoginByUsername)
	//
	// {
	//   "code": 1,
	//   "key": "USER::NOT_FOUND_BY_NAME",
	//   "title": "用户名不存在",
	//   "detail": "用户名: nobody 不存在",
	//   "errors": [
	//     {
	//       "code": 1,
	//       "key": "USER::NOT_FOUND_BY_NAME",
	//       "title": "用户名不存在",
	//       "detail": "用户名: nobody 不存在",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
	//
	// 404 1 USER::NOT_FOUND_BY_NAME 用户名不存在 用户名: nobody 不存在
}

// spec: 标识层; 本次出错的说明 / 没有说明
func ExampleLoginByUsername_sceneUserPasswordUnmatched() {
	_ = show(root, loginBy("1", "x"), &muser.LoginResult{})

	// Output:
	// HTTP/1.1 401 Unauthorized
	// Content-Length: 179
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(LoginByUsername)
	//
	// {
	//   "code": 5,
	//   "key": "USER::PASSWORD_UNMATCHED",
	//   "title": "登陆密码错误",
	//   "errors": [
	//     {
	//       "code": 5,
	//       "key": "USER::PASSWORD_UNMATCHED",
	//       "title": "登陆密码错误",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// STATUS__INTERNAL_SERVER_ERROR.Wrap(fmt.Errorf("some error"))
// spec: 状态码解析 / 显式声明状态码; 标识层 / 没有标识层; 内部信息不外泄
func ExampleLoginByUsername_sceneStatusInternalServerError() {
	_ = show(root, loginBy("3", "3"), &muser.LoginResult{})

	// Output:
	// HTTP/1.1 500 Internal Server Error
	// Content-Length: 37
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(LoginByUsername)
	//
	// {
	//   "errors": [
	//     {
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 错误位置 / 缺少必填的 body 字段. the decoded *status.Detail keeps the
// location and pointer of the entry (错误位置 / 下游错误位置被保留).
func ExampleLoginByUsername_sceneValidationMissingRequiredPassword() {
	err := show(root, raw(http.MethodPut, "/login", "application/json", `{"username":"1"}`), nil)
	d, _ := errors.AsType[*status.Detail](err)
	fmt.Println()
	fmt.Println(d.StatusCode(), d.Errors[0].Location, d.Errors[0].Pointer)

	// Output:
	// HTTP/1.1 400 Bad Request
	// Content-Length: 169
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(LoginByUsername)
	//
	// {
	//   "code": 7,
	//   "key": "VALIDATION::MISSING_REQUIRED",
	//   "errors": [
	//     {
	//       "code": 7,
	//       "key": "VALIDATION::MISSING_REQUIRED",
	//       "location": "body",
	//       "pointer": "/password",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
	//
	// 400 body /password
}

// spec: 校验的失败顺序 / 同一 body 内的多个错误
func ExampleLoginByUsername_sceneValidationInputTypeUsername() {
	_ = show(root, raw(http.MethodPut, "/login", "application/json", `{"username":1}`), nil)

	// Output:
	// HTTP/1.1 400 Bad Request
	// Content-Length: 233
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(LoginByUsername)
	//
	// {
	//   "code": 5,
	//   "key": "VALIDATION::INPUT_TYPE",
	//   "detail": "expect: string got: number",
	//   "errors": [
	//     {
	//       "code": 5,
	//       "key": "VALIDATION::INPUT_TYPE",
	//       "detail": "expect: string got: number",
	//       "location": "body",
	//       "pointer": "/username",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 错误位置 / 请求体不是合法 JSON
func ExampleLoginByUsername_sceneInvalidJSON() {
	_ = show(root, raw(http.MethodPut, "/login", "application/json", `{"username":`), nil)

	// Output:
	// HTTP/1.1 400 Bad Request
	// Content-Length: 55
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(LoginByUsername)
	//
	// {
	//   "errors": [
	//     {
	//       "location": "body",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 成功响应的状态码 / 返回值声明状态码 (WithResponseStatusCode(204)); 响应头与 cookie
func ExampleCookieRefresh() {
	_ = show(root, &login.CookieRefresh{Token: token("1")}, nil)

	// Output:
	// HTTP/1.1 204 No Content
	// Content-Length: 0
	// Server: example@v0.0.1(CookieRefresh)
	// Set-Cookie: token=<jwt>; Expires=Thu, 01 Jan 2026 01:00:00 GMT
}

// spec: 标识层 / 标识层没有提示; 本次出错的说明 / 没有说明 (cause `invalid token` not leaked)
func ExampleCookieRefresh_sceneUserFailedToRefreshTokenInvalidToken() {
	_ = show(root, &login.CookieRefresh{Token: "x"}, nil)

	// Output:
	// HTTP/1.1 500 Internal Server Error
	// Content-Length: 131
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(CookieRefresh)
	//
	// {
	//   "code": 6,
	//   "key": "USER::FAILED_TO_REFRESH_TOKEN",
	//   "errors": [
	//     {
	//       "code": 6,
	//       "key": "USER::FAILED_TO_REFRESH_TOKEN",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

func ExampleCookieRefresh_sceneUserFailedToRefreshTokenEmptyToken() {
	_ = show(root, &login.CookieRefresh{}, nil)

	// Output:
	// HTTP/1.1 500 Internal Server Error
	// Content-Length: 131
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(CookieRefresh)
	//
	// {
	//   "code": 6,
	//   "key": "USER::FAILED_TO_REFRESH_TOKEN",
	//   "errors": [
	//     {
	//       "code": 6,
	//       "key": "USER::FAILED_TO_REFRESH_TOKEN",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 状态码解析 / 隔一层包装仍保留状态码; 标识层; 本次出错的说明
func ExampleCookieRefresh_sceneUserNotFoundByID() {
	_ = show(root, &login.CookieRefresh{Token: token("9")}, nil)

	// Output:
	// HTTP/1.1 404 Not Found
	// Content-Length: 231
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(CookieRefresh)
	//
	// {
	//   "code": 2,
	//   "key": "USER::NOT_FOUND_BY_ID",
	//   "title": "用ID不存在",
	//   "detail": "用户ID: 9 不存在",
	//   "errors": [
	//     {
	//       "code": 2,
	//       "key": "USER::NOT_FOUND_BY_ID",
	//       "title": "用ID不存在",
	//       "detail": "用户ID: 9 不存在",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 错误的统一出口 / 前置 operator 返回错误; 标识层 / confjwt TOKEN_VALIDATION
func ExampleAuth_sceneConfjwtTokenValidationEmptyToken() {
	_ = show(users, listUsers("x"), &muser.ListResult{})

	// Output:
	// HTTP/1.1 401 Unauthorized
	// Content-Length: 177
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(ListUsers)
	//
	// {
	//   "code": 1,
	//   "key": "CONFJWT::TOKEN_VALIDATION",
	//   "detail": "got empty token",
	//   "errors": [
	//     {
	//       "code": 1,
	//       "key": "CONFJWT::TOKEN_VALIDATION",
	//       "detail": "got empty token",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// the jwt library cause `token is malformed` stays in Desc, not in the body
func ExampleAuth_sceneConfjwtTokenValidationInvalidToken() {
	_ = show(users, listUsers("x"), &muser.ListResult{}, httpx.Metadata{"Authorization": {"Bearer x"}})

	// Output:
	// HTTP/1.1 401 Unauthorized
	// Content-Length: 175
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(ListUsers)
	//
	// {
	//   "code": 1,
	//   "key": "CONFJWT::TOKEN_VALIDATION",
	//   "detail": "failed parsing",
	//   "errors": [
	//     {
	//       "code": 1,
	//       "key": "CONFJWT::TOKEN_VALIDATION",
	//       "detail": "failed parsing",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 错误的统一出口 / 前置 operator 返回错误
func ExampleCurrentUserInfo_sceneUserNotFoundByID() {
	_ = show(users, listUsers("x"), &muser.ListResult{}, bearer("9"))

	// Output:
	// HTTP/1.1 404 Not Found
	// Content-Length: 231
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(ListUsers)
	//
	// {
	//   "code": 2,
	//   "key": "USER::NOT_FOUND_BY_ID",
	//   "title": "用ID不存在",
	//   "detail": "用户ID: 9 不存在",
	//   "errors": [
	//     {
	//       "code": 2,
	//       "key": "USER::NOT_FOUND_BY_ID",
	//       "title": "用ID不存在",
	//       "detail": "用户ID: 9 不存在",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// codex.New(sqlx ROLLBACK)
// spec: 状态码解析 / codex 错误码声明状态码; 标识层 (no identifier, no hint); 内部信息不外泄
func ExampleCurrentUserInfo_sceneSqlerrorRollback() {
	_ = show(users, listUsers("x"), &muser.ListResult{}, bearer("3"))

	// Output:
	// HTTP/1.1 500 Internal Server Error
	// Content-Length: 55
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(ListUsers)
	//
	// {
	//   "code": 3,
	//   "errors": [
	//     {
	//       "code": 3,
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 成功响应的状态码 / 其他方法返回数据
func ExampleListUsers() {
	muser.Reset()
	_ = show(users, listUsers("管理员"), &muser.ListResult{}, bearer("1"))

	// Output:
	// HTTP/1.1 200 OK
	// Content-Length: 149
	// Content-Type: application/json; charset=utf-8
	// Server: example@v0.0.1(ListUsers)
	//
	// {
	//   "data": [
	//     {
	//       "id": 0,
	//       "userID": "0",
	//       "username": "0",
	//       "type": "管理员",
	//       "createdAt": "2026-01-01T00:00:00Z",
	//       "lastActiveAt": "2026-01-01T00:00:00Z"
	//     }
	//   ],
	//   "total": 3
	// }
}

// spec: 错误位置 / 缺少必填的 query 参数
func ExampleListUsers_sceneValidationMissingRequiredMemberType() {
	req := raw(http.MethodGet, "/user", "", "")
	req.Header.Set("Authorization", "Bearer "+token("1"))
	_ = show(users, req, nil)

	// Output:
	// HTTP/1.1 400 Bad Request
	// Content-Length: 172
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(ListUsers)
	//
	// {
	//   "code": 7,
	//   "key": "VALIDATION::MISSING_REQUIRED",
	//   "errors": [
	//     {
	//       "code": 7,
	//       "key": "VALIDATION::MISSING_REQUIRED",
	//       "location": "query",
	//       "pointer": "/memberType",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 错误位置 / 缺少必填的 query 参数 (类型错误)
func ExampleListUsers_sceneInvalidQueryCreatedAt() {
	req := raw(http.MethodGet, "/user?memberType=x&createdAt=yesterday", "", "")
	req.Header.Set("Authorization", "Bearer "+token("1"))
	_ = show(users, req, nil)

	// Output:
	// HTTP/1.1 400 Bad Request
	// Content-Length: 79
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(ListUsers)
	//
	// {
	//   "errors": [
	//     {
	//       "location": "query",
	//       "pointer": "/createdAt",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

func ExampleFetchByUsername() {
	_ = show(users, &user.FetchByUsername{Name: "0"}, &muser.Info{}, bearer("1"))

	// Output:
	// HTTP/1.1 200 OK
	// Content-Length: 128
	// Content-Type: application/json; charset=utf-8
	// Server: example@v0.0.1(FetchByUsername)
	//
	// {
	//   "id": 0,
	//   "userID": "0",
	//   "username": "0",
	//   "type": "管理员",
	//   "createdAt": "2026-01-01T00:00:00Z",
	//   "lastActiveAt": "2026-01-01T00:00:00Z"
	// }
}

func ExampleFetchByUsername_sceneUserNotFoundByName() {
	_ = show(users, &user.FetchByUsername{Name: "nobody"}, &muser.Info{}, bearer("1"))

	// Output:
	// HTTP/1.1 404 Not Found
	// Content-Length: 255
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(FetchByUsername)
	//
	// {
	//   "code": 1,
	//   "key": "USER::NOT_FOUND_BY_NAME",
	//   "title": "用户名不存在",
	//   "detail": "用户名: nobody 不存在",
	//   "errors": [
	//     {
	//       "code": 1,
	//       "key": "USER::NOT_FOUND_BY_NAME",
	//       "title": "用户名不存在",
	//       "detail": "用户名: nobody 不存在",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 状态码解析 / 显式声明状态码; 标识层 / 没有标识层; 内部信息不外泄
func ExampleFetchByUsername_sceneStatusInternalServerError() {
	_ = show(users, &user.FetchByUsername{Name: "3"}, &muser.Info{}, bearer("1"))

	// Output:
	// HTTP/1.1 500 Internal Server Error
	// Content-Length: 37
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(FetchByUsername)
	//
	// {
	//   "errors": [
	//     {
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// UserPermission returns ErrorResponse(fmt.Errorf("PermissionDenied"), WithResponseStatusCode(403),
// WithResponseTitle("无权操作"), WithResponseDetail(...)) as its value, and the
// operators after it are not executed: user 1 is still 普通会员.
// spec: 错误响应携带响应头与 cookie / 前置 operator 以返回值返回; 响应选项 / 普通错误设置问题内容
func ExampleUserPermission_sceneStatusForbidden() {
	muser.Reset()
	_ = show(users, &user.PromoteUserLevel{UserID: "1"}, &muser.PromotionResult{}, bearer("1"))

	out := &muser.Info{}
	_, _ = users.Do(ctx, &user.FetchByUsername{Name: "1"}, bearer("1")).Into(out)
	fmt.Println()
	fmt.Println(out.Type)

	// Output:
	// HTTP/1.1 403 Forbidden
	// Content-Length: 101
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(PromoteUserLevel)
	//
	// {
	//   "title": "无权操作",
	//   "detail": "需要管理员权限操作[]",
	//   "errors": [
	//     {
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
	//
	// 普通会员
}

func ExamplePromoteUserLevel() {
	muser.Reset()
	_ = show(users, &user.PromoteUserLevel{UserID: "1"}, &muser.PromotionResult{}, bearer("0"))

	// Output:
	// HTTP/1.1 200 OK
	// Content-Length: 58
	// Content-Type: application/json; charset=utf-8
	// Server: example@v0.0.1(PromoteUserLevel)
	//
	// {
	//   "userID": "1",
	//   "prev": "普通会员",
	//   "curr": "高级会员"
	// }
}

func ExamplePromoteUserLevel_sceneUserNotFoundByID() {
	_ = show(users, &user.PromoteUserLevel{UserID: "9"}, &muser.PromotionResult{}, bearer("0"))

	// Output:
	// HTTP/1.1 404 Not Found
	// Content-Length: 231
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(PromoteUserLevel)
	//
	// {
	//   "code": 2,
	//   "key": "USER::NOT_FOUND_BY_ID",
	//   "title": "用ID不存在",
	//   "detail": "用户ID: 9 不存在",
	//   "errors": [
	//     {
	//       "code": 2,
	//       "key": "USER::NOT_FOUND_BY_ID",
	//       "title": "用ID不存在",
	//       "detail": "用户ID: 9 不存在",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

func ExamplePromoteUserLevel_sceneSqlerrorRollback() {
	_ = show(users, &user.PromoteUserLevel{UserID: "3"}, &muser.PromotionResult{}, bearer("0"))

	// Output:
	// HTTP/1.1 500 Internal Server Error
	// Content-Length: 55
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(PromoteUserLevel)
	//
	// {
	//   "code": 3,
	//   "errors": [
	//     {
	//       "code": 3,
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

func ExampleDeleteUserByUserID_sceneStatusForbidden() {
	_ = show(users, &user.DeleteUserByUserID{UserID: 1}, nil, bearer("1"))

	// Output:
	// HTTP/1.1 403 Forbidden
	// Content-Length: 101
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(DeleteUserByUserID)
	//
	// {
	//   "title": "无权操作",
	//   "detail": "需要管理员权限操作[]",
	//   "errors": [
	//     {
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// fmt.Errorf("UnknownError")
// spec: 状态码解析 / 没有任何声明; 标识层 / 没有标识层
func ExampleDeleteUserByUserID_sceneUnknown() {
	_ = show(users, &user.DeleteUserByUserID{UserID: 1}, nil, bearer("0"))

	// Output:
	// HTTP/1.1 500 Internal Server Error
	// Content-Length: 37
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(DeleteUserByUserID)
	//
	// {
	//   "errors": [
	//     {
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 错误位置 / 路径参数
func ExampleDeleteUserByUserID_sceneInvalidPathUserID() {
	req := raw(http.MethodDelete, "/user/abc", "", "")
	req.Header.Set("Authorization", "Bearer "+token("0"))
	_ = show(users, req, nil)

	// Output:
	// HTTP/1.1 400 Bad Request
	// Content-Length: 75
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(DeleteUserByUserID)
	//
	// {
	//   "errors": [
	//     {
	//       "location": "path",
	//       "pointer": "/userID",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 标识层; 本次出错的说明 (codex.Errorf)
func ExampleUserRegister_sceneUserUsernameConflict() {
	muser.Reset()
	_ = show(root, register("1"), &muser.Info{})

	// Output:
	// HTTP/1.1 409 Conflict
	// Content-Length: 287
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(UserRegister)
	//
	// {
	//   "code": 4,
	//   "key": "USER::USERNAME_CONFLICT",
	//   "title": "用户名已注册",
	//   "detail": "用户名 [1] 已存在, 请换一个试试",
	//   "errors": [
	//     {
	//       "code": 4,
	//       "key": "USER::USERNAME_CONFLICT",
	//       "title": "用户名已注册",
	//       "detail": "用户名 [1] 已存在, 请换一个试试",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}

// spec: 成功响应的状态码 / POST 返回数据
func ExampleUserRegister() {
	muser.Reset()
	_ = show(root, register("reg-a"), &muser.Info{})

	// Output:
	// HTTP/1.1 201 Created
	// Content-Length: 135
	// Content-Type: application/json; charset=utf-8
	// Server: example@v0.0.1(UserRegister)
	//
	// {
	//   "id": 4,
	//   "userID": "4",
	//   "username": "reg-a",
	//   "type": "普通会员",
	//   "createdAt": "2026-01-01T00:00:00Z",
	//   "lastActiveAt": "2026-01-01T00:00:00Z"
	// }
}

// registration closes when there are more than 5 users
func ExampleUserRegister_sceneUserRegistrationNotOpen() {
	muser.Reset()
	for _, name := range []string{"reg-a", "reg-b", "reg-c"} {
		_, _ = root.Do(ctx, register(name)).Into(&muser.Info{})
	}
	_ = show(root, register("reg-z"), &muser.Info{})

	// Output:
	// HTTP/1.1 403 Forbidden
	// Content-Length: 227
	// Content-Type: application/problem+json
	// Server: example@v0.0.1(UserRegister)
	//
	// {
	//   "code": 3,
	//   "key": "USER::REGISTRATION_NOT_OPEN",
	//   "title": "注册未开放",
	//   "detail": "敬请期待",
	//   "errors": [
	//     {
	//       "code": 3,
	//       "key": "USER::REGISTRATION_NOT_OPEN",
	//       "title": "注册未开放",
	//       "detail": "敬请期待",
	//       "src": "example@v0.0.1"
	//     }
	//   ]
	// }
}
