package rule

import (
	"bytes"
)

type Parameter interface {
	Node

	// Parameters returns parameter of rule
	// eg:
	//	@map<@string,@int?> => map value with string key and int value
	//	@string<byte>[10] => string value with max 10 byte length
	//	@string<rune>[10] => string value with max 10 rune length
	//	@slice<@float64<10,4>[-1000.0001,1000.0002]> => slice with float64 with width 10 and precision 4 and between -1000.0001 and 1000.0002
	Parameters() []Node
	AddParameters(...Node)
}

func NewParameters() Parameter {
	return &parameters{NodeType: NODE_TYPE__PARAMETERS}
}

type parameters struct {
	NodeType

	parameters []Node
}

func (ps *parameters) IsNil() bool {
	return ps == nil || len(ps.parameters) == 0
}

func (ps *parameters) Parameters() []Node {
	return ps.parameters
}

func (ps *parameters) AddParameters(nodes ...Node) {
	for _, node := range nodes {
		if !IsNil(node) {
			ps.parameters = append(ps.parameters, node)
		}
	}
}

func (ps *parameters) Bytes() []byte {
	if IsNil(ps) {
		return nil
	}

	b := bytes.NewBuffer(nil)
	b.WriteRune(TOK_PARAM_L)
	for i, p := range ps.parameters {
		if i > 0 {
			b.WriteRune(TOK_SPLITTER)
		}
		b.Write(p.Bytes())
	}
	b.WriteRune(TOK_PARAM_R)

	return b.Bytes()
}
