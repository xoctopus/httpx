package rule

type Literal struct {
	NodeType

	data []byte
}

func NewLiteral(text string) *Literal {
	return &Literal{
		NodeType: NODE_TYPE__LITERAL,
		data:     []byte(text),
	}
}

func (l *Literal) IsNil() bool {
	return l == nil || len(l.data) == 0
}

func (l *Literal) Append(b ...byte) {
	l.data = append(l.data, b...)
}

func (l *Literal) String() string {
	return string(l.Bytes())
}

func (l *Literal) Bytes() []byte {
	if IsNil(l) {
		return nil
	}
	return l.data
}
