package status_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/validation"
)

func TestResponse(t *testing.T) {
	bdd.From(t).Given("a request located error", func(t bdd.T) {
		err := validation.PrefixPointer(validation.WrapPointer(errors.New("0"), "/name"), "/body")
		r := status.AsResponse(err, "srv")

		t.When("resolved", func(t bdd.T) {
			t.Then("400 for located error", bdd.Equal(r.StatusCode(), http.StatusBadRequest))
			t.Then("status text", bdd.Equal(r.StatusText(), http.StatusText(http.StatusBadRequest)))
			t.Then("no identity layer", bdd.Equal(r.Key, ""))
			t.Then("one entry", bdd.HaveLen(r.Errors, 1))
			t.Then("entry pointer", bdd.Equal(string(r.Errors[0].Pointer), "/body/name"))
			t.Then("cause not in detail", bdd.Equal(r.Errors[0].Detail, ""))
			t.Then("unwrap entries", bdd.HaveLen(r.Unwrap(), 1))
		})
		t.When("status overridden", func(t bdd.T) {
			r.SetStatusCode(http.StatusUnprocessableEntity)
			t.Then("status", bdd.Equal(r.StatusCode(), http.StatusUnprocessableEntity))
		})
	})

	bdd.From(t).Given("an empty response", func(t bdd.T) {
		t.When("unwrap", func(t bdd.T) {
			t.Then("no entries", bdd.HaveLen((&status.Response{}).Unwrap(), 0))
		})
	})
}

func TestDetailStatus(t *testing.T) {
	bdd.From(t).Given("downstream sub problems", func(t bdd.T) {
		d := status.UnmarshalResponse(http.StatusBadRequest, []byte(`{"key":"X","errors":[{"key":"Y"}]}`))
		t.When("status of sub problem", func(t bdd.T) {
			var de status.Describer
			t.Then("is describer", bdd.BeTrue(errors.As(error(d.Errors[0]), &de)))
			t.Then("unspecified", bdd.Equal(de.StatusCode(), 0))
			t.Then("no status text", bdd.Equal(d.Errors[0].StatusText(), ""))
		})
	})

	bdd.From(t).Given("non-standard status code", func(t bdd.T) {
		d := status.UnmarshalResponse(499, nil)
		t.When("unmarshalled", func(t bdd.T) {
			t.Then("no key", bdd.Equal(d.Key, ""))
			t.Then("status", bdd.Equal(d.StatusCode(), 499))
		})
	})

	bdd.From(t).Given("src set by caller", func(t bdd.T) {
		x := &status.Detail{Src: "downstream.host"}
		t.When("unmarshalled", func(t bdd.T) {
			t.Then("succeed", bdd.Succeed(x.UnmarshalResponse(http.StatusBadRequest, []byte(`{"key":"X"}`))))
			t.Then("src kept", bdd.Equal(x.Src, "downstream.host"))
			t.Then("key", bdd.Equal(x.Key, "X"))
		})
	})
}
