package status_test

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/status"
)

// expect is the expected top level of a Response, zero values mean absent
type expect struct {
	status int
	code   int64
	key    string
	title  string
	detail string
}

func expectResponse(t bdd.T, r *status.Response, e expect) {
	t.Then("status", bdd.Equal(r.StatusCode(), e.status))
	t.Then("code", bdd.Equal(r.Code, e.code))
	t.Then("key", bdd.Equal(r.Key, e.key))
	t.Then("title", bdd.Equal(r.Title, e.title))
	t.Then("detail", bdd.Equal(r.Detail, e.detail))
}

func marshal(r *status.Response) (string, map[string]any) {
	data, _ := json.Marshal(r)
	m := map[string]any{}
	_ = json.Unmarshal(data, &m)
	return string(data), m
}

// TestAsResponse covers spec: 状态码解析, 标识层, 本次出错的说明
func TestAsResponse(t *testing.T) {
	sqlNotFound := codex.New(SQL__NOTFOUND)

	bdd.From(t).Given("status code resolution", func(t bdd.T) {
		t.When("codex code declares status", func(t bdd.T) {
			expectResponse(t, status.AsResponse(codex.New(USER__NO_HINT), "srv"),
				expect{status: 401, code: 3, key: "USER::NO_HINT"})
		})
		t.When("wrapped one more layer", func(t bdd.T) {
			r := status.AsResponse(fmt.Errorf("svc: %w", codex.New(USER__NOT_FOUND)), "srv")
			expectResponse(t, r, expect{status: 404, code: 1, key: "USER::NOT_FOUND", title: "用户不存在"})
		})
		t.When("status declared explicitly on a plain error", func(t bdd.T) {
			expectResponse(t, status.AsResponse(status.WrapCode(errors.New("some error"), 500), "srv"),
				expect{status: 500})
		})
		t.When("status overridden explicitly", func(t bdd.T) {
			r := status.AsResponse(status.WrapCode(codex.New(USER__NOT_FOUND), http.StatusForbidden), "srv")
			expectResponse(t, r, expect{status: 403, code: 1, key: "USER::NOT_FOUND", title: "用户不存在"})
		})
		t.When("outer code declares no status", func(t bdd.T) {
			r := status.AsResponse(codex.Wrap(USER__QUERY_FAILED, sqlNotFound), "srv")
			expectResponse(t, r, expect{status: 404, code: 2, key: "USER::QUERY_FAILED", title: "查询失败"})
		})
		t.When("nothing declared", func(t bdd.T) {
			expectResponse(t, status.AsResponse(errors.New("UnknownError"), "srv"), expect{status: 500})
		})
	})

	bdd.From(t).Given("identity layer", func(t bdd.T) {
		t.When("domain code wraps component code", func(t bdd.T) {
			r := status.AsResponse(codex.Wrapf(USER__NOT_FOUND, sqlNotFound, "user %d not found", 9), "srv")
			expectResponse(t, r, expect{
				status: 404, code: 1, key: "USER::NOT_FOUND", title: "用户不存在", detail: "user 9 not found",
			})
		})
		t.When("identity layer has no hint", func(t bdd.T) {
			r := status.AsResponse(codex.Wrap(USER__NO_HINT, codex.New(USER__NOT_FOUND)), "srv")
			t.Then("no title, not taken from inner", bdd.Equal(r.Title, ""))
		})
		t.When("identity layer has no identifier", func(t bdd.T) {
			expectResponse(t, status.AsResponse(codex.New(SQL__ROLLBACK), "srv"), expect{status: 500, code: 2})
		})
		t.When("no identity layer", func(t bdd.T) {
			expectResponse(t, status.AsResponse(status.WrapCode(fmt.Errorf("some error"), 503), "srv"),
				expect{status: 503})
		})
	})

	bdd.From(t).Given("detail of this occurrence", func(t bdd.T) {
		t.When("codex.Errorf", func(t bdd.T) {
			r := status.AsResponse(codex.Errorf(USER__NOT_FOUND, "id %s", "9"), "srv")
			t.Then("detail formatted", bdd.Equal(r.Detail, "id 9"))
		})
		t.When("codex.New", func(t bdd.T) {
			t.Then("no detail", bdd.Equal(status.AsResponse(codex.New(USER__NOT_FOUND), "srv").Detail, ""))
		})
	})

	bdd.From(t).Given("errors.Join", func(t bdd.T) {
		e1, e2 := codex.New(USER__NOT_FOUND), codex.New(USER__NO_HINT)
		t.When("outer has identity layer", func(t bdd.T) {
			r := status.AsResponse(codex.Wrap(USER__QUERY_FAILED, errors.Join(e1, e2)), "srv")
			t.Then("top from outer", bdd.Equal(r.Key, "USER::QUERY_FAILED"))
			t.Then("entries per branch", bdd.HaveLen(r.Errors, 2))
			t.Then("branch resolves own identity", bdd.Equal(r.Errors[1].Key, "USER::NO_HINT"))
		})
		t.When("outer has no identity layer", func(t bdd.T) {
			r := status.AsResponse(errors.Join(e1, e2), "srv")
			t.Then("top from first branch", bdd.Equal(r.Key, "USER::NOT_FOUND"))
			t.Then("status from first branch", bdd.Equal(r.StatusCode(), 404))
		})
	})

	bdd.From(t).Given("nil error", func(t bdd.T) {
		t.When("resolve", func(t bdd.T) {
			t.Then("nil response", bdd.BeNil(status.AsResponse(nil, "srv")))
		})
	})
}

// TestResponseBody covers spec: 错误响应体结构, 内部信息不外泄
func TestResponseBody(t *testing.T) {
	bdd.From(t).Given("domain code wraps a cause", func(t bdd.T) {
		cause := errors.New("dial tcp 10.0.0.1:3306: refused")
		r := status.AsResponse(codex.Wrapf(USER__NOT_FOUND, cause, "user %d not found", 9), "srv")
		data, body := marshal(r)

		t.When("serialized", func(t bdd.T) {
			keys := slices.Sorted(maps.Keys(body))
			t.Then("top level fields", bdd.Equal(keys, []string{"code", "detail", "errors", "key", "title"}))
			t.Then("cause not leaked", bdd.BeFalse(strings.Contains(data, "dial tcp")))
			t.Then("message not leaked", bdd.BeFalse(strings.Contains(data, "[USER:1]")))
			t.Then("no desc", bdd.BeFalse(strings.Contains(data, `"desc"`)))
		})
		t.When("logging", func(t bdd.T) {
			t.Then("desc keeps the full text", bdd.ContainsSubString(r.Desc, "dial tcp"))
			t.Then("entry desc keeps the full text", bdd.ContainsSubString(r.Errors[0].Desc, "dial tcp"))
		})
	})

	bdd.From(t).Given("plain error", func(t bdd.T) {
		r := status.AsResponse(errors.New("UnknownError"), "srv")
		data, body := marshal(r)
		t.When("serialized", func(t bdd.T) {
			t.Then("only errors at top level", bdd.Equal(slices.Sorted(maps.Keys(body)), []string{"errors"}))
			t.Then("text not leaked", bdd.BeFalse(strings.Contains(data, "UnknownError")))
			for _, f := range []string{`"type"`, `"status"`, `"instance"`} {
				t.Then("no "+f, bdd.BeFalse(strings.Contains(data, f)))
			}
		})
	})
}

// TestAsDetail covers a single problem and downstream details
func TestAsDetail(t *testing.T) {
	bdd.From(t).Given("a codex error", func(t bdd.T) {
		d := status.AsDetail(codex.Errorf(USER__NOT_FOUND, "id %s", "9"), "srv")
		t.When("resolved", func(t bdd.T) {
			t.Then("status", bdd.Equal(d.StatusCode(), 404))
			t.Then("key", bdd.Equal(d.Key, "USER::NOT_FOUND"))
			t.Then("detail", bdd.Equal(d.Detail, "id 9"))
			t.Then("src", bdd.Equal(d.Src, "srv"))
		})
	})

	bdd.From(t).Given("a downstream detail", func(t bdd.T) {
		down := &status.Detail{Key: "DOWN::X", Status: 404}
		t.When("resolved", func(t bdd.T) {
			t.Then("kept as is", bdd.Equal(status.AsDetail(fmt.Errorf("w: %w", down), "srv"), down))
		})
	})
}
