package rule

import (
	"bytes"
)

type Matrix interface {
	Node

	// ValueMatrix returns fields enumerated values
	// eg: @int{1,2,3} => int value must in {1,2,3}
	ValueMatrix() [][]*Literal
	AppendValues(...[]*Literal)
}

func NewMatrix() Matrix {
	return &matrix{NodeType: NODE_TYPE__ENUMERATIONS}
}

type matrix struct {
	NodeType

	matrix [][]*Literal
}

func (vs *matrix) IsNil() bool {
	return vs == nil || len(vs.matrix) == 0
}

func (vs *matrix) ValueMatrix() [][]*Literal {
	if !IsNil(vs) {
		return vs.matrix
	}
	return nil
}

func (vs *matrix) AppendValues(values ...[]*Literal) {
	vs.matrix = append(vs.matrix, values...)
}

func (vs *matrix) Bytes() []byte {
	if IsNil(vs) {
		return nil
	}

	b := bytes.NewBuffer(nil)
	for _, mx := range vs.matrix {
		b.WriteRune(TOK_VALUES_L)
		for i, v := range mx {
			if i > 0 {
				b.WriteRune(TOK_SPLITTER)
			}
			b.Write(v.Bytes())
		}
		b.WriteRune(TOK_VALUES_R)
	}
	return b.Bytes()
}
