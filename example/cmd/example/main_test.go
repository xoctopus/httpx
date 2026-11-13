package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/logx"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/example/cmd/example/apis"
	"github.com/xoctopus/httpx/example/cmd/example/apis/login"
	"github.com/xoctopus/httpx/example/cmd/example/apis/user"
	muser "github.com/xoctopus/httpx/example/pkg/modules/user"
	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/pkg/confhttp"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/midx"
	"github.com/xoctopus/httpx/pkg/routex"
)

// Acceptance tests of the response-model spec (openspec/changes/define-response-model),
// driven through the example service: requests are sent by confhttp.Client and
// responses are decoded by Result.Into, as a downstream service would do.
//
// For every case:
//   - entry: an operator value sent by Client.Do, or a raw *http.Request for inputs
//     an operator value cannot express
//   - exit:  Result.Into; a non 2xx response is returned as *status.Detail
//   - expect: the raw response (status line, headers, body) is checked strictly
//     against the spec; the decoded *status.Detail is checked on the status code
//     and on the fields the body carries
//
// The cases share the in-memory user store of the example and run in order.

var (
	jwt = &confjwt.JWT{Issuer: "example", ExpIn: types.Duration(time.Hour), SignKey: "x"}

	endpoint string           // http://127.0.0.1:{port}/example/v0
	root     *confhttp.Client // {endpoint}
	users    *confhttp.Client // {endpoint}/user
)

func TestMain(m *testing.M) {
	ctx := logx.With(context.Background(), logx.NewStd())

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
		if r := call(root, &apis.HealthCheck{}, nil); r.Err == nil && r.Status == http.StatusNoContent {
			break
		}
		must.BeTrue(i < 100)
		time.Sleep(20 * time.Millisecond)
	}

	os.Exit(m.Run())
}

// recorder keeps the raw response while Result.Into still reads the body
type recorder struct {
	status int
	header http.Header
	body   []byte
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	rsp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	r.status, r.header = rsp.StatusCode, rsp.Header
	r.body, err = io.ReadAll(rsp.Body)
	_ = rsp.Body.Close()
	rsp.Body = io.NopCloser(bytes.NewReader(r.body))
	return rsp, err
}

type result struct {
	Status int
	Header http.Header
	Raw    []byte
	Body   map[string]any // raw body decoded if it is a JSON object
	Err    error          // returned by Result.Into
}

// call sends req (an operator value or *http.Request) and decodes the response into out
func call(c *confhttp.Client, req any, out any, metas ...httpx.Metadata) *result {
	rec := &recorder{}
	ctx := confhttp.WithHttpClient(context.Background(), &http.Client{Transport: rec})

	_, err := c.Do(ctx, req, metas...).Into(out)

	r := &result{Status: rec.status, Header: rec.header, Raw: rec.body, Err: err}
	if len(r.Raw) > 0 && r.Raw[0] == '{' {
		_ = json.Unmarshal(r.Raw, &r.Body)
	}
	return r
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

func has(m map[string]any, key string) bool { _, ok := m[key]; return ok }

func str(m map[string]any, key string) string { s, _ := m[key].(string); return s }

// problem is the expected top level of an error response; zero values mean the
// field MUST be absent.
type problem struct {
	Status int
	Code   int64
	Key    string
	Title  string
	Detail string
}

// ExpectSuccessNoContent checks a success response without body, spec: 成功响应的状态码
func (r *result) ExpectSuccessNoContent(t bdd.T) {
	t.Then("decoded without error", bdd.Succeed(r.Err))
	t.Then("status 204", bdd.Equal(r.Status, http.StatusNoContent))
	t.Then("empty body", bdd.HaveLen(r.Raw, 0))
}

// ExpectSuccessContent checks a success response with a JSON body decoded into
// the output, spec: 成功响应的状态码, Content-Type
func (r *result) ExpectSuccessContent(t bdd.T, code int) {
	t.Then("decoded without error", bdd.Succeed(r.Err))
	t.Then("status", bdd.Equal(r.Status, code))
	t.Then("json body", bdd.HavePrefix(r.Header.Get("Content-Type"), "application/json"))
	t.Then("non empty body", bdd.BeTrue(len(r.Raw) > 0))
}

// ExpectProblem checks an error response against p
func (r *result) ExpectProblem(t bdd.T, p problem) {
	// raw response, spec: 错误响应的媒体类型, 错误响应体结构, 标识层, 本次出错的说明, 内部信息不外泄
	t.Then("status", bdd.Equal(r.Status, p.Status))
	t.Then("problem media type", bdd.Equal(r.Header.Get("Content-Type"), "application/problem+json"))
	if p.Code == 0 {
		t.Then("body: no code", bdd.BeFalse(has(r.Body, "code")))
	} else {
		code, _ := r.Body["code"].(float64)
		t.Then("body: code", bdd.Equal(int64(code), p.Code))
	}
	for _, f := range []struct{ name, expect string }{{"key", p.Key}, {"title", p.Title}, {"detail", p.Detail}} {
		if f.expect == "" {
			t.Then("body: no "+f.name, bdd.BeFalse(has(r.Body, f.name)))
		} else {
			t.Then("body: "+f.name, bdd.Equal(str(r.Body, f.name), f.expect))
		}
	}
	for _, f := range []string{"type", "status", "instance", "src", "desc"} {
		t.Then("body: top level has no "+f, bdd.BeFalse(has(r.Body, f)))
	}

	// decoded by Result.Into
	d, ok := errors.AsType[*status.Detail](r.Err)
	t.Then("decoded as *status.Detail", bdd.BeTrue(ok))
	if !ok {
		return
	}
	t.Then("decoded: status code", bdd.Equal(d.StatusCode(), p.Status))
	if p.Code != 0 {
		t.Then("decoded: code", bdd.Equal(d.Code, p.Code))
	}
	if p.Key != "" {
		t.Then("decoded: key", bdd.Equal(d.Key, p.Key))
	}
	if p.Title != "" {
		t.Then("decoded: title", bdd.Equal(d.Title, p.Title))
	}
	if p.Detail != "" {
		t.Then("decoded: detail", bdd.Equal(d.Detail, p.Detail))
	}
}

// ExpectInvalid checks a request parameter error, spec: 错误的统一出口 / 参数校验失败.
// error position (loc / pos) is not covered in this change.
func (r *result) ExpectInvalid(t bdd.T) {
	t.Then("status 400", bdd.Equal(r.Status, http.StatusBadRequest))
	t.Then("problem media type", bdd.Equal(r.Header.Get("Content-Type"), "application/problem+json"))
	for _, f := range []string{"type", "status", "instance", "src", "desc"} {
		t.Then("body: top level has no "+f, bdd.BeFalse(has(r.Body, f)))
	}

	d, ok := errors.AsType[*status.Detail](r.Err)
	t.Then("decoded as *status.Detail", bdd.BeTrue(ok))
	if ok {
		t.Then("decoded: status code", bdd.Equal(d.StatusCode(), http.StatusBadRequest))
	}
}

// ExpectNotLeaked checks the body does not contain internal texts
func (r *result) ExpectNotLeaked(t bdd.T, texts ...string) {
	for _, text := range texts {
		t.Then("body does not contain "+text, bdd.BeFalse(bytes.Contains(r.Raw, []byte(text))))
	}
}

var (
	userNotFoundByID = problem{
		Status: 404, Code: 2, Key: "USER::NOT_FOUND_BY_ID", Title: "用ID不存在", Detail: "用户ID: 9 不存在",
	}
	// UserPermission returns ErrorResponse(fmt.Errorf("PermissionDenied"), WithResponseStatusCode(403),
	// WithResponseTitle("无权操作"), WithResponseDetail(...)) as its value
	permissionDenied = problem{Status: 403, Title: "无权操作", Detail: "需要管理员权限操作[]"}
	// confjwt.Auth returns CONFJWT::TOKEN_VALIDATION (401) for empty / invalid Bearer
	tokenEmpty   = problem{Status: 401, Code: 1, Key: "CONFJWT::TOKEN_VALIDATION", Detail: "got empty token"}
	tokenInvalid = problem{Status: 401, Code: 1, Key: "CONFJWT::TOKEN_VALIDATION", Detail: "failed parsing"}
)

func TestHealthCheck(t *testing.T) {
	bdd.From(t).Given("service is up", func(t bdd.T) {
		t.When("GET /health-check", func(t bdd.T) {
			r := call(root, &apis.HealthCheck{}, nil)
			// spec: 成功响应的状态码 / 返回 nil
			r.ExpectSuccessNoContent(t)
		})
	})
}

func TestLoginByUsername(t *testing.T) {
	body := func(username, password string) *login.LoginByUsername {
		r := &login.LoginByUsername{}
		r.Username, r.Password = username, password
		return r
	}

	bdd.From(t).Given("username and password match", func(t bdd.T) {
		t.When("PUT /login", func(t bdd.T) {
			out := &muser.LoginResult{}
			r := call(root, body("1", "1"), out)
			// spec: 成功响应的状态码 / 其他方法返回数据; Content-Type
			r.ExpectSuccessContent(t, http.StatusOK)
			t.Then("decoded into result", bdd.Equal(out.UserID, "1"))
		})
	})

	bdd.From(t).Given("username does not exist", func(t bdd.T) {
		t.When("PUT /login", func(t bdd.T) {
			r := call(root, body("nobody", "x"), &muser.LoginResult{})
			// spec: 标识层 / 领域错误包装组件错误; 本次出错的说明 / 带说明的领域错误
			r.ExpectProblem(t, problem{
				Status: 404, Code: 1, Key: "USER::NOT_FOUND_BY_NAME", Title: "用户名不存在", Detail: "用户名: nobody 不存在",
			})
			r.ExpectNotLeaked(t, "SQLERROR", "NOTFOUND")
		})
	})

	bdd.From(t).Given("password does not match", func(t bdd.T) {
		t.When("PUT /login", func(t bdd.T) {
			r := call(root, body("1", "x"), &muser.LoginResult{})
			// spec: 标识层; 本次出错的说明 / 没有说明
			r.ExpectProblem(t, problem{Status: 401, Code: 5, Key: "USER::PASSWORD_UNMATCHED", Title: "登陆密码错误"})
		})
	})

	bdd.From(t).Given("lookup fails", func(t bdd.T) {
		t.When("PUT /login with username 3", func(t bdd.T) {
			r := call(root, body("3", "3"), &muser.LoginResult{})
			// STATUS__INTERNAL_SERVER_ERROR.Wrap(fmt.Errorf("some error"))
			// spec: 状态码解析 / 显式声明状态码; 标识层 / 没有标识层; 内部信息不外泄
			r.ExpectProblem(t, problem{Status: 500})
			r.ExpectNotLeaked(t, "some error")
		})
	})

	bdd.From(t).Given("body is not valid json", func(t bdd.T) {
		t.When("PUT /login", func(t bdd.T) {
			r := call(root, raw(http.MethodPut, "/login", "application/json", `{"username":`), nil)
			r.ExpectInvalid(t)
		})
	})
}

func TestCookieRefresh(t *testing.T) {
	refresh := func(tok string) *login.CookieRefresh { return &login.CookieRefresh{Token: tok} }

	bdd.From(t).Given("token cookie is valid", func(t bdd.T) {
		t.When("PUT /token-refresh", func(t bdd.T) {
			r := call(root, refresh(token("1")), nil)
			// spec: 成功响应的状态码 / 返回值声明状态码 (WithResponseStatusCode(204)); 响应头与 cookie
			r.ExpectSuccessNoContent(t)
			t.Then("token cookie refreshed", bdd.ContainsSubString(r.Header.Get("Set-Cookie"), "token="))
		})
	})

	bdd.From(t).Given("token cookie is invalid", func(t bdd.T) {
		t.When("PUT /token-refresh", func(t bdd.T) {
			r := call(root, refresh("x"), nil)
			// spec: 标识层 / 标识层没有提示; 本次出错的说明 / 没有说明 (cause not leaked)
			r.ExpectProblem(t, problem{Status: 500, Code: 6, Key: "USER::FAILED_TO_REFRESH_TOKEN"})
			r.ExpectNotLeaked(t, "invalid token")
		})
	})

	bdd.From(t).Given("token cookie is absent", func(t bdd.T) {
		t.When("PUT /token-refresh", func(t bdd.T) {
			r := call(root, refresh(""), nil)
			r.ExpectProblem(t, problem{Status: 500, Code: 6, Key: "USER::FAILED_TO_REFRESH_TOKEN"})
		})
	})

	bdd.From(t).Given("token belongs to a missing user", func(t bdd.T) {
		t.When("PUT /token-refresh", func(t bdd.T) {
			r := call(root, refresh(token("9")), nil)
			// spec: 状态码解析 / 隔一层包装仍保留状态码; 标识层; 本次出错的说明
			r.ExpectProblem(t, userNotFoundByID)
		})
	})
}

func TestAuth(t *testing.T) {
	list := &user.ListUsers{}
	list.MemberType = "x"

	bdd.From(t).Given("authorization is absent", func(t bdd.T) {
		t.When("GET /user", func(t bdd.T) {
			r := call(users, list, &muser.ListResult{})
			// spec: 错误的统一出口 / 前置 operator 返回错误; 标识层 / confjwt TOKEN_VALIDATION
			r.ExpectProblem(t, tokenEmpty)
		})
	})

	bdd.From(t).Given("token is invalid", func(t bdd.T) {
		t.When("GET /user", func(t bdd.T) {
			r := call(users, list, &muser.ListResult{}, httpx.Metadata{"Authorization": {"Bearer x"}})
			r.ExpectProblem(t, tokenInvalid)
			// jwt library cause stays in Desc, not in problem body
			r.ExpectNotLeaked(t, "token is malformed")
		})
	})

	bdd.From(t).Given("token belongs to a missing user", func(t bdd.T) {
		t.When("GET /user", func(t bdd.T) {
			r := call(users, list, &muser.ListResult{}, bearer("9"))
			// spec: 错误的统一出口 / 前置 operator 返回错误 (raised by CurrentUserInfo)
			r.ExpectProblem(t, userNotFoundByID)
		})
	})

	bdd.From(t).Given("storage fails when loading current user", func(t bdd.T) {
		t.When("GET /user", func(t bdd.T) {
			r := call(users, list, &muser.ListResult{}, bearer("3"))
			// codex.New(sqlx ROLLBACK): spec: 状态码解析 / codex 错误码声明状态码; 标识层 (no identifier, no hint)
			r.ExpectProblem(t, problem{Status: 500, Code: 3})
			r.ExpectNotLeaked(t, "SQLERROR", "ROLLBACK")
		})
	})
}

func TestListUsers(t *testing.T) {
	bdd.From(t).Given("authorized", func(t bdd.T) {
		t.When("query is valid", func(t bdd.T) {
			req := &user.ListUsers{}
			req.MemberType = "管理员"
			out := &muser.ListResult{}
			r := call(users, req, out, bearer("1"))
			// spec: 成功响应的状态码 / 其他方法返回数据
			r.ExpectSuccessContent(t, http.StatusOK)
			t.Then("decoded into result", bdd.BeTrue(out.Total > 0))
		})
		t.When("required memberType is absent", func(t bdd.T) {
			req := raw(http.MethodGet, "/user", "", "")
			req.Header.Set("Authorization", "Bearer "+token("1"))
			call(users, req, nil).ExpectInvalid(t)
		})
		t.When("createdAt is not a time", func(t bdd.T) {
			req := raw(http.MethodGet, "/user?memberType=x&createdAt=yesterday", "", "")
			req.Header.Set("Authorization", "Bearer "+token("1"))
			call(users, req, nil).ExpectInvalid(t)
		})
	})
}

func TestFetchByUsername(t *testing.T) {
	bdd.From(t).Given("authorized", func(t bdd.T) {
		t.When("user exists", func(t bdd.T) {
			out := &muser.Info{}
			r := call(users, &user.FetchByUsername{Name: "0"}, out, bearer("1"))
			r.ExpectSuccessContent(t, http.StatusOK)
			t.Then("decoded into result", bdd.Equal(out.Username, "0"))
		})
		t.When("user does not exist", func(t bdd.T) {
			r := call(users, &user.FetchByUsername{Name: "nobody"}, &muser.Info{}, bearer("1"))
			r.ExpectProblem(t, problem{
				Status: 404, Code: 1, Key: "USER::NOT_FOUND_BY_NAME", Title: "用户名不存在", Detail: "用户名: nobody 不存在",
			})
		})
		t.When("lookup fails", func(t bdd.T) {
			r := call(users, &user.FetchByUsername{Name: "3"}, &muser.Info{}, bearer("1"))
			// spec: 状态码解析 / 显式声明状态码; 标识层 / 没有标识层; 内部信息不外泄
			r.ExpectProblem(t, problem{Status: 500})
			r.ExpectNotLeaked(t, "some error")
		})
	})
}

func TestPromoteUserLevel(t *testing.T) {
	bdd.From(t).Given("current user has no permission", func(t bdd.T) {
		t.When("PUT /user/promote/1", func(t bdd.T) {
			r := call(users, &user.PromoteUserLevel{UserID: "1"}, &muser.PromotionResult{}, bearer("1"))
			// spec: 错误响应携带响应头与 cookie / 前置 operator 以返回值返回; 响应选项 / 普通错误设置问题内容
			r.ExpectProblem(t, permissionDenied)
			r.ExpectNotLeaked(t, "PermissionDenied")
		})
		t.When("the api after the permission check", func(t bdd.T) {
			out := &muser.Info{}
			call(users, &user.FetchByUsername{Name: "1"}, out, bearer("1"))
			t.Then("is not executed", bdd.Equal(out.Type, "普通会员"))
		})
	})

	bdd.From(t).Given("current user is admin", func(t bdd.T) {
		t.When("user exists", func(t bdd.T) {
			out := &muser.PromotionResult{}
			r := call(users, &user.PromoteUserLevel{UserID: "2"}, out, bearer("0"))
			r.ExpectSuccessContent(t, http.StatusOK)
			t.Then("decoded into result", bdd.Equal(out.Curr, "高级会员"))
		})
		t.When("user does not exist", func(t bdd.T) {
			r := call(users, &user.PromoteUserLevel{UserID: "9"}, &muser.PromotionResult{}, bearer("0"))
			r.ExpectProblem(t, userNotFoundByID)
		})
		t.When("storage fails", func(t bdd.T) {
			r := call(users, &user.PromoteUserLevel{UserID: "3"}, &muser.PromotionResult{}, bearer("0"))
			r.ExpectProblem(t, problem{Status: 500, Code: 3})
		})
	})
}

func TestDeleteUserByUserID(t *testing.T) {
	bdd.From(t).Given("current user has no permission", func(t bdd.T) {
		t.When("DELETE /user/1", func(t bdd.T) {
			r := call(users, &user.DeleteUserByUserID{UserID: 1}, nil, bearer("1"))
			r.ExpectProblem(t, permissionDenied)
		})
	})

	bdd.From(t).Given("current user is admin", func(t bdd.T) {
		t.When("DELETE /user/1", func(t bdd.T) {
			r := call(users, &user.DeleteUserByUserID{UserID: 1}, nil, bearer("0"))
			// spec: 状态码解析 / 没有任何声明; 标识层 / 没有标识层 (fmt.Errorf("UnknownError"))
			r.ExpectProblem(t, problem{Status: 500})
			r.ExpectNotLeaked(t, "UnknownError")
		})
		t.When("userID is not an integer", func(t bdd.T) {
			req := raw(http.MethodDelete, "/user/abc", "", "")
			req.Header.Set("Authorization", "Bearer "+token("0"))
			call(users, req, nil).ExpectInvalid(t)
		})
	})
}

// TestUserRegister runs last: it fills the user store until registration closes
// (more than 5 users).
func TestUserRegister(t *testing.T) {
	register := func(username string) *login.UserRegister {
		return &login.UserRegister{Body: muser.RegisterParam{Username: username, Password: "x"}}
	}

	bdd.From(t).Given("username is taken", func(t bdd.T) {
		t.When("POST /register", func(t bdd.T) {
			r := call(root, register("1"), &muser.Info{})
			// spec: 标识层; 本次出错的说明 (codex.Errorf)
			r.ExpectProblem(t, problem{
				Status: 409, Code: 4, Key: "USER::USERNAME_CONFLICT", Title: "用户名已注册",
				Detail: "用户名 [1] 已存在, 请换一个试试",
			})
		})
	})

	bdd.From(t).Given("username is new", func(t bdd.T) {
		t.When("POST /register", func(t bdd.T) {
			out := &muser.Info{}
			r := call(root, register("reg-a"), out)
			// spec: 成功响应的状态码 / POST 返回数据
			r.ExpectSuccessContent(t, http.StatusCreated)
			t.Then("decoded into result", bdd.Equal(out.Username, "reg-a"))
		})
	})

	bdd.From(t).Given("more than 5 users", func(t bdd.T) {
		for _, name := range []string{"reg-b", "reg-c", "reg-d"} {
			call(root, register(name), &muser.Info{})
		}
		t.When("POST /register", func(t bdd.T) {
			r := call(root, register("reg-z"), &muser.Info{})
			r.ExpectProblem(t, problem{
				Status: 403, Code: 3, Key: "USER::REGISTRATION_NOT_OPEN", Title: "注册未开放", Detail: "敬请期待",
			})
		})
	})
}
