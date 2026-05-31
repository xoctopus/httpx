package validation

import (
	"reflect"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/validation/rule"
)

func wrap(v Validator, r rule.Rule) Validator {
	optional, defaults := true, []byte("")
	if r != nil {
		optional, defaults = r.Optional(), r.Defaults().Bytes()
	}
	return &wrapped{
		Validator: v,
		optional:  optional,
		defaults:  defaults,
	}
}

type wrapped struct {
	Validator
	optional bool
	defaults []byte
}

func (o *wrapped) String() string {
	if o.Validator != nil {
		text := o.Validator.String()
		if o.optional {
			return text + "?"
		}
		return text
	}
	return ""
}

func (o *wrapped) PostValidate(rv reflect.Value) error {
	if post, ok := o.Validator.(PostValidator); ok {
		return post.PostValidate(rv)
	}
	return nil
}

func (o *wrapped) ElemRule() rule.Rule {
	if x, ok := o.Validator.(WithElemRule); ok {
		return x.ElemRule()
	}
	return nil
}

func (o *wrapped) KeyRule() rule.Rule {
	if x, ok := o.Validator.(WithKeyRule); ok {
		return x.KeyRule()
	}
	return nil
}

func (o *wrapped) Validate(value []byte) error {
	if jsontext.Value(value).Kind() == jsontext.NULL && !o.optional {
		return codex.New(ERROR__MISSING_REQUIRED)
	}
	if o.Validator == nil {
		return nil
	}
	return o.Validator.Validate(value)
}

func (o *wrapped) Defaults() []byte {
	return o.defaults
}

func (o *wrapped) Optional() bool {
	return o.optional
}
