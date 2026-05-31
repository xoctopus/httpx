package validation

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/xoctopus/x/codex"
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

type WithKeyRule interface {
	KeyRule() rule.Rule
}

type WithElemRule interface {
	ElemRule() rule.Rule
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
	rules:     syncx.NewXmap[Option, func() (Validator, error)](),
}

func Register(p Provider) {
	_, ok := gValidators.providers[p.Name()]
	if ok {
		// must.BeTrueF(!ok, "%s validator not be registered", p.Name())
		// fmt.Printf("WARN %s validator have been registered", p.Name())
		return
	}
	gValidators.providers[p.Name()] = p

	for _, name := range p.Variants() {
		if _, ok = gValidators.providers[name]; !ok {
			gValidators.providers[name] = p
		}
	}
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
		if o.Rule == nil || o.Rule.IsNil() {
			if text := DefaultRuleByType(o.Type); len(text) > 0 {
				r, err := rule.CompileAsBuilder(text)
				if err != nil {
					return nil, err
				}
				if o.Rule != nil {
					r.SetOptional(o.Rule.Optional())
					r.SetDefaults(o.Rule.Defaults())
				}
				o.Rule = r
			}
		}

		// no rule no validator
		if o.Rule == nil || o.Rule.IsNil() {
			return nil, nil
		}

		name := o.Rule.Name()
		p, ok := vs.providers[name]
		if !ok {
			return nil, codex.Errorf(ERROR__UNREGISTERED_RULE, "rule name: %s", o.Rule.Name())
		}
		v, err := p.New(o.Rule)
		if err != nil {
			return nil, err
		}
		return wrap(v, o.Rule), nil
	}))
	return get()
}

func DefaultRuleByType(t reflect.Type) string {
	if t == nil {
		return ""
	}

	if t.Implements(reflect.TypeFor[TagValidator]()) {
		return reflect.New(t).Interface().(TagValidator).ValidationTag()
	}

	if t.Implements(tTextUnmarshaler) || reflect.PointerTo(t).Implements(tTextUnmarshaler) {
		return "@string?"
	}

	switch t {
	case tBytes:
		return "@string?"
	default:
		switch t.Kind() {
		case reflect.Pointer:
			return DefaultRuleByType(t.Elem())
		case reflect.Array:
			if elem := DefaultRuleByType(t.Elem()); len(elem) > 0 {
				return fmt.Sprintf("@slice<%s>[%d]?", elem, t.Len())
			}
			return fmt.Sprintf("@slice[%d]?", t.Len())
		case reflect.Slice:
			if elem := DefaultRuleByType(t.Elem()); len(elem) > 0 {
				return fmt.Sprintf("@slice<%s>?", elem)
			}
			return "@slice?"
		case reflect.Map:
			kr, vr := DefaultRuleByType(t.Key()), DefaultRuleByType(t.Elem())
			if len(kr) == 0 && len(vr) == 0 {
				return "@map?"
			}
			return fmt.Sprintf("@map<%s,%s>?", kr, vr)
		case reflect.Bool:
			return "@bool"
		case reflect.Int:
			return "@int"
		case reflect.Int8:
			return "@int8"
		case reflect.Int16:
			return "@int16"
		case reflect.Int32:
			return "@int32"
		case reflect.Int64:
			return "@int64"
		case reflect.Uint:
			return "@uint"
		case reflect.Uint8:
			return "@uint8"
		case reflect.Uint16:
			return "@uint16"
		case reflect.Uint32:
			return "@uint32"
		case reflect.Uint64:
			return "@uint64"
		case reflect.Float32:
			return "@float32"
		case reflect.Float64:
			return "@float64"
		case reflect.String:
			return "@string"
		default:
			return "" // nil validator is allowed
		}
	}
}
