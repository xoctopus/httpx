package validators

import (
	"regexp"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
)

func NewUserDefinedProvider(name string, fn func(unquoted string) error) validation.Provider {
	p := &_userDefinedP{
		name: name,
		fn:   fn,
	}
	validation.Register(p)
	return p
}

func NewRegexpProvider(exp, name, hint string) validation.Provider {
	re := regexp.MustCompile(exp)

	return NewUserDefinedProvider(
		name,
		func(s string) error {
			if re.MatchString(s) {
				return nil
			}
			return codex.Errorf(validation.ERROR__NOT_MATCH_REGEXP, "expect: %s - (%s) got %s", exp, hint, s)
		},
	)
}

type _userDefinedP struct {
	name string
	fn   func(unquoted string) error
}

func (p *_userDefinedP) Name() string {
	return p.name
}

func (p *_userDefinedP) Variants() []string {
	return nil
}

func (p *_userDefinedP) New(r rule.Rule) (validation.Validator, error) {
	return &UserDefined{
		name: r.Name(),
		f:    p.fn,
	}, nil
}

type UserDefined struct {
	name string
	f    func(unquoted string) error
}

func (v *UserDefined) Format() string {
	return v.name
}

func (v *UserDefined) String() string {
	return "@" + v.name
}

func (v *UserDefined) Validate(value []byte) error {
	if k := jsontext.Value(value).Kind(); k != jsontext.STRING {
		return codex.Errorf(validation.ERROR__INPUT_TYPE, "expect: string got: %s", k)
	}

	unquote, err := jsontext.AppendUnquote(nil, value)
	if err != nil {
		return err
	}
	val := string(unquote)

	if err = v.f(val); err != nil {
		return err
	}

	return nil
}
