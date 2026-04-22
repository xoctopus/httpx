package validators

import (
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
)

func init() {
	validation.Register(&_booleanP{})
}

type _booleanP struct{}

func (_booleanP) Name() string {
	return "bool"
}

func (_booleanP) Variants() []string {
	return []string{"bool", "boolean"}
}

func (*_booleanP) New(_ rule.Rule) (validation.Validator, error) {
	return &Boolean{}, nil
}

type Boolean struct{}

func (v *Boolean) String() string {
	return "@bool"
}

func (v *Boolean) Validate(value []byte) error {
	k := jsontext.Value(value).Kind()
	if k != jsontext.FALSE && k != jsontext.TRUE {
		return codex.Errorf(validation.ERROR__INPUT_TYPE, "expect: bool got: %s", k)
	}
	return nil
}
