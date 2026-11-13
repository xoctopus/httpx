package status_test

import (
	"encoding/json/jsontext"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/status"
)

// located is an error carrying a request location and optionally a position
type located struct {
	error
	location string
	pointer  jsontext.Pointer
}

func (e *located) Location() string          { return e.location }
func (e *located) Pointer() jsontext.Pointer { return e.pointer }
func (e *located) Unwrap() error             { return e.error }

// TestLocationStatus covers spec: 位置决定状态码兜底
func TestLocationStatus(t *testing.T) {
	bdd.From(t).Given("an error with a location and no declared status", func(t bdd.T) {
		r := status.AsResponse(&located{error: errors.New("x"), location: "body"}, "srv")
		t.When("resolved", func(t bdd.T) {
			// spec: 位置决定状态码兜底 / 只有 location 的错误
			t.Then("400", bdd.Equal(r.StatusCode(), http.StatusBadRequest))
		})
	})

	bdd.From(t).Given("an error with a location and a declared status", func(t bdd.T) {
		err := status.WrapCode(&located{error: errors.New("x"), location: "body"}, http.StatusInternalServerError)
		r := status.AsResponse(err, "srv")
		t.When("resolved", func(t bdd.T) {
			// spec: 位置决定状态码兜底 / 已声明状态码不被兜底覆盖
			t.Then("declared status wins", bdd.Equal(r.StatusCode(), http.StatusInternalServerError))
		})
	})

	bdd.From(t).Given("an error without a location and a declared status", func(t bdd.T) {
		r := status.AsResponse(errors.New("x"), "srv")
		t.When("resolved", func(t bdd.T) {
			// spec: 位置决定状态码兜底 / 没有声明也没有位置
			t.Then("500", bdd.Equal(r.StatusCode(), http.StatusInternalServerError))
		})
	})
}

// TestLocationBody covers spec: 错误位置 / 非解析错误没有位置, 没有 pointer
func TestLocationBody(t *testing.T) {
	bdd.From(t).Given("an error with a location and a position", func(t bdd.T) {
		r := status.AsResponse(&located{error: errors.New("x"), location: "body", pointer: jsontext.Pointer("/password")}, "srv")
		data, _ := marshal(r)
		t.When("resolved", func(t bdd.T) {
			t.Then("entry location", bdd.Equal(r.Errors[0].Location, "body"))
			t.Then("entry pointer", bdd.Equal(string(r.Errors[0].Pointer), "/password"))
			t.Then("pointer member", bdd.BeTrue(strings.Contains(data, `"pointer":"/password"`)))
			t.Then("top level has no location", bdd.BeFalse(strings.HasPrefix(data, `{"location"`)))
		})
	})

	bdd.From(t).Given("an error with a location only", func(t bdd.T) {
		r := status.AsResponse(&located{error: errors.New("x"), location: "body"}, "srv")
		data, _ := marshal(r)
		t.When("resolved", func(t bdd.T) {
			t.Then("entry location", bdd.Equal(r.Errors[0].Location, "body"))
			t.Then("no pointer member", bdd.BeFalse(strings.Contains(data, `"pointer"`)))
		})
	})

	bdd.From(t).Given("an error without a location", func(t bdd.T) {
		r := status.AsResponse(errors.New("x"), "srv")
		data, _ := marshal(r)
		t.When("resolved", func(t bdd.T) {
			t.Then("no location member", bdd.BeFalse(strings.Contains(data, `"location"`)))
			t.Then("no pointer member", bdd.BeFalse(strings.Contains(data, `"pointer"`)))
		})
	})
}

// TestUnmarshalLocation covers spec: 错误位置 / 下游错误位置被保留
func TestUnmarshalLocation(t *testing.T) {
	bdd.From(t).Given("a downstream body with positions", func(t bdd.T) {
		d := status.UnmarshalResponse(http.StatusBadRequest, []byte(
			`{"key":"X","errors":[{"key":"Y","location":"body","pointer":"/a/b"},{"location":"query"}]}`,
		))
		t.When("unmarshalled", func(t bdd.T) {
			t.Then("two entries", bdd.HaveLen(d.Errors, 2))
			t.Then("location", bdd.Equal(d.Errors[0].Location, "body"))
			t.Then("pointer", bdd.Equal(string(d.Errors[0].Pointer), "/a/b"))
			t.Then("location only", bdd.Equal(d.Errors[1].Location, "query"))
			t.Then("no pointer", bdd.Equal(string(d.Errors[1].Pointer), ""))
		})
	})
}
