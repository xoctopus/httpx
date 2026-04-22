package validators

import (
	"reflect"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
)

func wrap(v validation.Validator, r rule.Rule) validation.Validator {
	return &wrapped{
		Validator: v,
		optional:  r.Optional(),
		defaults:  r.Defaults().Bytes(),
	}
}

type wrapped struct {
	validation.Validator
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
	if post, ok := o.Validator.(validation.PostValidator); ok {
		return post.PostValidate(rv)
	}
	return nil
}

func (o *wrapped) Elem() validation.Option {
	if post, ok := o.Validator.(validation.WithElem); ok {
		return post.Elem()
	}
	return validation.Option{}
}

func (o *wrapped) Key() validation.Option {
	if post, ok := o.Validator.(validation.WithKey); ok {
		return post.Key()
	}
	return validation.Option{}
}

func (o *wrapped) Validate(value []byte) error {
	if jsontext.Value(value).Kind() == jsontext.NULL && !o.optional {
		return codex.New(validation.ERROR__MISSING_REQUIRED)
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
