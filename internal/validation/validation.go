package validation

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/syncx"

	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/internal/validation/rule"
)

type Provider interface {
	Name() string
	Variants() []string

	New(rule.Rule) (Validator, error)
}

type Validator interface {
	String() string
	Validate([]byte) error
}

type PostValidator interface {
	PostValidate(reflect.Value) error
}

type WithOptional interface {
	Optional() bool
}

type WithDefaults interface {
	Defaults() []byte
}

type WithKey interface {
	Key() Option
}

// WithElem element option of map
type WithElem interface {
	Elem() Option
}

type TagValidator interface {
	ValidationTag() string
}

type Option struct {
	Type reflect.Type
	Rule rule.Rule
	Text bool
}

var gValidators = &validators{
	providers: make(map[string]Provider),
}

func Register(p Provider) {
	_, ok := gValidators.providers[p.Name()]
	must.BeTrueF(!ok, "%s validator not be registered")
	gValidators.providers[p.Name()] = p

	for _, name := range p.Variants() {
		if _, ok = gValidators.providers[name]; !ok {
			gValidators.providers[p.Name()] = p
		}
	}
	return
}

func New(option Option) (Validator, error) {
	return gValidators.New(option)
}

func NewFromStructField(f *scanner.Field) (Validator, error) {
	opt := Option{
		Type: f.Type,
		Text: f.String,
	}

	b := rule.NewBuilder("")

	if tag, ok := f.Tag.Lookup("validate"); ok {
		r, err := rule.Compile(tag)
		if err != nil {
			return nil, err
		}
		b = r.(rule.Builder)
	}

	if v, ok := f.Tag.Lookup("default"); ok {
		b.SetDefaults(rule.NewLiteral(v))
	}
	b.SetOptional(f.Omitzero || f.Omitempty)
	opt.Rule = b

	return New(opt)
}

type validators struct {
	providers map[string]Provider
	rules     syncx.Map[Option, func() (Validator, error)]
}

func (vs *validators) New(o Option) (Validator, error) {
	get, _ := vs.rules.LoadOrStore(o, sync.OnceValues(func() (Validator, error) {
		if o.Type != nil {
			if v, ok := reflect.New(o.Type).Interface().(TagValidator); ok {
				r, err := rule.Compile(v.ValidationTag())
				if err != nil {
					return nil, fmt.Errorf("failed to compile validation tag: %s %v", v.ValidationTag(), err)
				}
				o.Rule = r
			}
		}

		p, ok := vs.providers[o.Rule.Name()]
		if !ok {
			return nil, fmt.Errorf("unsupported rule: %s", o.Rule.Name())
		}
		return p.New(o.Rule)
	}))
	return get()
}
