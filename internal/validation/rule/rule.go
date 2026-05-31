package rule

import (
	"bytes"
	"regexp"
	"text/scanner"
	"unicode"

	"github.com/xoctopus/x/misc/must"
)

// Compile compiles validation rule
// TODO cache compiled rules for acceleration
func Compile[T ~[]byte | ~string](data T) (Rule, error) {
	s := &scanner.Scanner{
		IsIdentRune: func(ch rune, i int) bool {
			return unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0 || ch == '-' && i > 0
		},
	}
	s.Init(bytes.NewReader([]byte(data)))
	return fromScanner(&_scanner{Scanner: s, data: []byte(data)})
}

func CompileAsBuilder[T ~[]byte | ~string](data T) (Builder, error) {
	r, err := Compile(data)
	if err != nil {
		return nil, err
	}
	return r.(Builder), nil
}

func MustCompile[T ~[]byte | ~string](data T) Rule {
	return must.NoErrorV(Compile(data))
}

func fromScanner(s *_scanner) (Rule, error) {
	r := &rule{NodeType: NODE_TYPE__ROOT}
	if err := r.init(s); err != nil {
		return nil, err
	}
	return r, nil
}

// NewBuilder creates validation rule builder
func NewBuilder(name string) Builder {
	return &rule{NodeType: NODE_TYPE__ROOT, name: NewLiteral(name)}
}

type Rule interface {
	Node

	// Name returns name of rule
	Name() string
	// Parameters returns parameter of rule
	// eg:
	//	@map<@string,@int?> => map value with string key and int value
	//	@string<byte>[10] => string value with max 10 byte length
	//	@string<rune>[10] => string value with max 10 rune length
	//	@slice<@float64<10,4>[-1000.0001,1000.0002]> => slice with float64 with width 10 and precision 4 and between -1000.0001 and 1000.0002
	Parameters() []Node
	// Min returns range minimum value and exclusive
	// eg: @int(0,10) => min: 0 exclusive: true
	Min() (*Literal, bool)
	// Max returns range maximum value and exclusive
	// eg: @int(0,10] => max: 10 exclusive: false
	Max() (*Literal, bool)
	// LengthMode return if range use LengthMode
	// eg: @string[10] => length of string must be 10
	LengthMode() bool
	// ValueMatrix returns value's enumeration value matrix
	// eg:
	//	@string{A,B,C} value must be A, B or C
	//	@slice<@int>{1,2,3}{4,5,6} slice value must be []int{1,2,3} or []int{4,5,6}
	ValueMatrix() [][]*Literal
	// Pattern returns regexp literal
	// eg: @string='s'/\w+/ => string value must match \w+
	Pattern() string
	// Regexp returns compiled regexp
	Regexp() *regexp.Regexp
	// Optional returns if field is optional
	// eg: @string? => optional string value
	Optional() bool
	// Defaults returns default value of field
	// eg: @string='abc' => string value defaults is "abc" if not assigned
	Defaults() *Literal
}

type Builder interface {
	Node

	// Name returns name of rule
	Name() string
	// SetName set rule name
	SetName(string)

	Parameter
	Ranges
	Matrix
	Matcher
	Requirements
}

func (r *rule) init(s *_scanner) (err error) {
	var pos, end int
	// rule name
	r.name, pos, err = s.name()
	if err != nil {
		return err
	}

LOOP:
	for tok := s.Peek(); ; tok = s.Peek() {
		switch tok {
		case ' ':
			s.Next()
		case TOK_PARAM_L:
			r.Parameter, err = s.parameters()
		case TOK_RANGE_IN_L, TOK_RANGE_EX_L:
			r.Ranges, err = s.ranges()
		case TOK_VALUES_L:
			var values []*Literal
			values, err = s.values()
			if err == nil {
				if r.Matrix == nil {
					r.Matrix = NewMatrix()
				}
				r.Matrix.AppendValues(values)
			}
		case TOK_REGEXP:
			r.Matcher, err = s.pattern()
		case TOK_EQUAL, TOK_OPTIONAL:
			r.Requirements, err = s.requirements()
		default:
			break LOOP
		}
		if err != nil {
			return
		}
	}

	end = s.Pos().Offset
	r.data = s.data[pos:end]

	return nil
}

// rule validation rule
// @name?<parameters>[range_or_length]
type rule struct {
	NodeType

	data []byte
	name *Literal

	Parameter
	Ranges
	Matrix
	Matcher
	Requirements
}

func (r *rule) Type() NodeType {
	return r.NodeType.Type()
}

func (r *rule) IsNil() bool {
	return r == nil || IsNil(r.name)
}

func (r *rule) Name() string {
	return r.name.String()
}

func (r *rule) SetName(name string) {
	r.name = NewLiteral(name)
}

func (r *rule) Parameters() []Node {
	if !IsNil(r.Parameter) {
		return r.Parameter.Parameters()
	}
	return nil
}

func (r *rule) AddParameters(nodes ...Node) {
	if r.Parameter == nil {
		r.Parameter = NewParameters()
	}
	r.Parameter.AddParameters(nodes...)
}

func (r *rule) Min() (*Literal, bool) {
	if r.Ranges == nil {
		return nil, false
	}
	return r.Ranges.Min()
}

func (r *rule) SetMin(l *Literal, ex bool) {
	if r.Ranges == nil {
		r.Ranges = NewRanges()
	}
	r.Ranges.SetMin(l, ex)
}

func (r *rule) Max() (*Literal, bool) {
	if r.Ranges == nil {
		return nil, false
	}
	return r.Ranges.Max()
}

func (r *rule) SetMax(l *Literal, ex bool) {
	if r.Ranges == nil {
		r.Ranges = NewRanges()
	}
	r.Ranges.SetMax(l, ex)
}

func (r *rule) LengthMode() bool {
	return r.Ranges != nil && r.Ranges.LengthMode()
}

func (r *rule) SetLengthMode(b bool) {
	if r.Ranges == nil {
		r.Ranges = NewRanges()
	}
	r.Ranges.SetLengthMode(b)
}

func (r *rule) ValueMatrix() [][]*Literal {
	if r.Matrix == nil {
		return nil
	}
	return r.Matrix.ValueMatrix()
}

func (r *rule) AppendValues(values ...[]*Literal) {
	if r.Matrix == nil {
		r.Matrix = NewMatrix()
	}
	r.Matrix.AppendValues(values...)
}

func (r *rule) Pattern() string {
	if r.Matcher == nil {
		return ""
	}
	return r.Matcher.Pattern()
}

func (r *rule) Regexp() *regexp.Regexp {
	if r.Matcher == nil {
		return nil
	}
	return r.Matcher.Regexp()
}

func (r *rule) SetPattern(pattern string) error {
	if r.Matcher == nil {
		r.Matcher = NewMatcher()
	}
	return r.Matcher.SetPattern(pattern)
}

func (r *rule) Optional() bool {
	return r.Requirements != nil && r.Requirements.Optional()
}

func (r *rule) SetOptional(b bool) {
	if r.Requirements == nil {
		r.Requirements = NewRequirements()
	}
	r.Requirements.SetOptional(b)
}

func (r *rule) Defaults() *Literal {
	if r.Requirements == nil {
		return nil
	}
	return r.Requirements.Defaults()
}

func (r *rule) SetDefaults(l *Literal) {
	if r.Requirements == nil {
		r.Requirements = NewRequirements()
	}
	r.Requirements.SetDefaults(l)
}

func (r *rule) Bytes() []byte {
	b := bytes.NewBuffer(nil)

	b.WriteRune(TOK_LEADER)
	b.Write(r.name.Bytes())

	if !IsNil(r.Parameter) {
		b.Write(r.Parameter.Bytes())
	}

	if !IsNil(r.Ranges) {
		b.Write(r.Ranges.Bytes())
	}

	if !IsNil(r.Matrix) {
		b.Write(r.Matrix.Bytes())
	}

	if !IsNil(r.Matcher) {
		b.Write(r.Matcher.Bytes())
	}

	if !IsNil(r.Requirements) {
		b.Write(r.Requirements.Bytes())
	}

	if len(r.data) == 0 {
		r.data = b.Bytes()
	}

	return b.Bytes()
}
