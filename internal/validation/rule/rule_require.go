package rule

import (
	"bytes"
)

type Requirements interface {
	Node

	// Optional returns if field is optional
	// eg: @string? => optional string value
	Optional() bool
	SetOptional(bool)

	// Defaults returns default value of field
	// eg: @string='abc' => string value defaults is "abc" if not assigned
	Defaults() *Literal
	SetDefaults(*Literal)
}

func NewRequirements() Requirements {
	return &requirements{NodeType: NODE_TYPE__REQUIREMENTS}
}

type requirements struct {
	NodeType

	optional bool
	defaults *Literal
}

func (a *requirements) IsNil() bool {
	return a == nil
}

func (a *requirements) Optional() bool {
	return a.optional || !IsNil(a.defaults)
}

func (a *requirements) SetOptional(b bool) {
	a.optional = b
}

func (a *requirements) Defaults() *Literal {
	return a.defaults
}

func (a *requirements) SetDefaults(l *Literal) {
	a.defaults = l
}

func (a *requirements) Bytes() []byte {
	if a.Optional() {
		b := bytes.NewBuffer(nil)
		if !IsNil(a.defaults) {
			b.WriteRune(TOK_EQUAL)
			b.Write(SingleQuote(a.defaults.Bytes()))
		} else {
			b.WriteRune(TOK_OPTIONAL)
		}
		return b.Bytes()
	}
	return nil
}
