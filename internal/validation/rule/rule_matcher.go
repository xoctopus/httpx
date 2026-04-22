package rule

import (
	"regexp"

	"github.com/xoctopus/x/codex"
)

type Matcher interface {
	Node

	// Pattern returns regexp literal
	// eg: @string='s'/\w+/ => string value must match \w+
	Pattern() string
	SetPattern(string) error
	// Regexp returns compiled regexp
	Regexp() *regexp.Regexp
}

func NewMatcher() Matcher {
	return &matcher{NodeType: NODE_TYPE__REGEXP_MATCHER}
}

type matcher struct {
	NodeType

	pattern string
	regexp  *regexp.Regexp
}

func (m *matcher) IsNil() bool {
	return m == nil || len(m.pattern) == 0 || m.regexp == nil
}

func (m *matcher) Pattern() string {
	return m.pattern
}

func (m *matcher) SetPattern(pattern string) error {
	m.pattern, m.regexp = pattern, nil
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return codex.Wrapf(ERROR__FAILED_COMPILE_REGEXP, err, "pattern: %s", m.Pattern())
	}
	m.regexp = regex
	return nil
}

func (m *matcher) Regexp() *regexp.Regexp {
	return m.regexp
}

func (m *matcher) Bytes() []byte {
	return Slash([]byte(m.pattern))
}
