package rule

import (
	"bytes"
)

// Ranges describes validation value's range
// eg:
//
//	@int32[,10]   => int32 value must greater than or equal to 0 and less than or equal to 10
//	@int64(,10)   => int64 value must greater than 0 and less or equal than 10
//	@string[10,)  => string length must greater than or equal to 10
//	@string[10]   => string length must equal to 10
type Ranges interface {
	Node

	Min() (*Literal, bool)
	SetMin(*Literal, bool)
	Max() (*Literal, bool)
	SetMax(*Literal, bool)

	LengthMode() bool
	SetLengthMode(bool)
}

func NewRanges() Ranges {
	return &ranges{NodeType: NODE_TYPE__RANGE}
}

type ranges struct {
	NodeType

	lengthMode bool
	min        *Literal
	exclusiveL bool
	max        *Literal
	exclusiveR bool
}

func (r *ranges) IsNil() bool {
	return r == nil || IsNil(r.min) && IsNil(r.max)
}

func (r *ranges) Min() (*Literal, bool) {
	if !IsNil(r) {
		return r.min, r.exclusiveL
	}
	return nil, false
}

func (r *ranges) SetMin(l *Literal, ex bool) {
	r.min, r.exclusiveL = l, ex
}

func (r *ranges) Max() (*Literal, bool) {
	if !IsNil(r) {
		return r.max, r.exclusiveR
	}
	return nil, false
}

func (r *ranges) SetMax(l *Literal, ex bool) {
	r.max, r.exclusiveR = l, ex
}

func (r *ranges) LengthMode() bool {
	return r.lengthMode
}

func (r *ranges) SetLengthMode(b bool) {
	r.lengthMode = b
}

func (r *ranges) Bytes() []byte {
	if IsNil(r) {
		return nil
	}

	b := bytes.NewBuffer(nil)

	if r.LengthMode() {
		b.WriteRune(TOK_RANGE_IN_L)
		b.Write(r.min.Bytes())
		b.WriteRune(TOK_RANGE_IN_R)
		return b.Bytes()
	}

	if r.exclusiveL {
		b.WriteRune(TOK_RANGE_EX_L)
	} else {
		b.WriteRune(TOK_RANGE_IN_L)
	}

	b.Write(r.min.Bytes())
	b.WriteRune(TOK_SPLITTER)
	b.Write(r.max.Bytes())

	if r.exclusiveR {
		b.WriteRune(TOK_RANGE_EX_R)
	} else {
		b.WriteRune(TOK_RANGE_IN_R)
	}
	return b.Bytes()
}
