package rule

type Node interface {
	// Type presents rule node type
	Type() NodeType
	// IsNil returns if node is nil
	IsNil() bool
	// Bytes returns formatted node content
	Bytes() []byte
}

// NodeType rule node type
// +genx:enum
type NodeType int8

const (
	NODE_TYPE_UNKNOWN NodeType = iota
	NODE_TYPE__ROOT   NodeType = iota + 0
	NODE_TYPE__PARAMETERS
	NODE_TYPE__RANGE
	NODE_TYPE__ENUMERATIONS
	NODE_TYPE__REGEXP_MATCHER
	NODE_TYPE__REQUIREMENTS
	_ // main node add above, fake node add below
	NODE_TYPE__LITERAL
)

func (t NodeType) Type() NodeType {
	return t
}

func IsNil(n Node) bool {
	return n == nil || n.IsNil()
}
