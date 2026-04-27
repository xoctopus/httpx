package rule

import (
	"bytes"
	"text/scanner"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/misc/must"
)

const (
	TOK_LEADER     rune = '@'
	TOK_OPTIONAL   rune = '?'
	TOK_EQUAL      rune = '='
	TOK_RANGE_IN_L rune = '['
	TOK_RANGE_IN_R rune = ']'
	TOK_RANGE_EX_L rune = '('
	TOK_RANGE_EX_R rune = ')'
	TOK_VALUES_L   rune = '{'
	TOK_VALUES_R   rune = '}'
	TOK_PARAM_L    rune = '<'
	TOK_PARAM_R    rune = '>'
	TOK_REGEXP     rune = '/'
	TOK_SPLITTER   rune = ','
)

var KeyTokens = map[rune]struct{}{
	TOK_LEADER:     {},
	TOK_OPTIONAL:   {},
	TOK_EQUAL:      {},
	TOK_RANGE_IN_L: {},
	TOK_RANGE_IN_R: {},
	TOK_RANGE_EX_L: {},
	TOK_RANGE_EX_R: {},
	TOK_VALUES_L:   {},
	TOK_VALUES_R:   {},
	TOK_PARAM_L:    {},
	TOK_PARAM_R:    {},
	TOK_REGEXP:     {},
	TOK_SPLITTER:   {},
}

type _scanner struct {
	data []byte
	*scanner.Scanner
}

func (s *_scanner) literal() (string, error) {
	tok := s.Scan()
	if _, ok := KeyTokens[tok]; ok {
		return "", codex.Errorf(
			ERROR__INVALID_LITERAL,
			"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
		)
	}
	return s.TokenText(), nil
}

// name scans rule name.
// eg: @string @int @email
func (s *_scanner) name() (*Literal, int, error) {
	if tok := s.Next(); tok != TOK_LEADER {
		return nil, 0, codex.Errorf(
			ERROR__INVALID_RULE_LEADER,
			"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
		)
	}
	pos := s.Pos().Offset - 1
	name, err := s.literal()
	if err != nil {
		return nil, 0, err
	}
	if len(name) == 0 {
		return nil, 0, codex.New(ERROR__MISSING_RULE_NAME)
	}
	return NewLiteral(name), pos, nil
}

func (s *_scanner) parameters() (_ Parameter, err error) {
	tok := s.Next()
	must.BeTrueWrap(
		tok == TOK_PARAM_L,
		codex.Errorf(
			ERROR__INVALID_PARAM_LEADER,
			"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
		),
	)

	var (
		params = NewParameters()
		curr   Node
	)

	for tok = s.Peek(); tok != TOK_PARAM_R; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, codex.Errorf(
				ERROR__INVALID_PARAM_TAILER,
				"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
			)
		}
		switch tok {
		case ' ':
			s.Next()
		case TOK_SPLITTER:
			params.AddParameters(curr)
			curr = nil
			s.Next()
		case TOK_LEADER:
			curr, err = fromScanner(s)
			if err != nil {
				return
			}
		default:
			var text string
			text, err = s.literal()
			if err != nil {
				return nil, err
			}
			if curr == nil {
				params.AddParameters(NewLiteral(text))
				continue
			}
			if lit, ok := curr.(*Literal); ok {
				lit.Append([]byte(text)...)
				continue
			}
			return nil, codex.Errorf(
				ERROR__INVALID_PARAMS,
				"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
			)
		}
	}
	params.AddParameters(curr)
	s.Next()
	return params, nil
}

func (s *_scanner) ranges() (rs Ranges, err error) {
	if tok := s.Peek(); tok != TOK_RANGE_EX_L && tok != TOK_RANGE_IN_L {
		return nil, codex.Errorf(
			ERROR__INVALID_RANGE_LEADER,
			"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
		)
	}

	var (
		list   = make([]*Literal, 0, 2)
		leader = s.Next()
		tailer rune
		curr   *Literal
	)

LOOP:
	for tok := s.Peek(); ; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, codex.Errorf(
				ERROR__INVALID_RANGE_TAILER,
				"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
			)
		}
		switch tok {
		case ' ':
			s.Next()
			continue
		case TOK_SPLITTER:
			s.Next()
			list = append(list, curr)
			curr = nil
		case TOK_RANGE_EX_R, TOK_RANGE_IN_R:
			tailer = s.Next()
			list = append(list, curr)
			break LOOP
		default:
			var text string
			text, err = s.literal()
			if err != nil {
				return nil, err
			}
			if curr == nil {
				curr = NewLiteral(text)
				continue
			}
			curr.Append(byte(tok))
		}
	}
	rs = NewRanges()
	mode := len(list) == 1 && leader == TOK_RANGE_IN_L && tailer == TOK_RANGE_IN_R

	switch len(list) {
	case 0:
		return nil, nil
	case 1:
		if IsNil(list[0]) {
			return nil, nil
		}
		if !mode {
			return nil, codex.Errorf(ERROR__INVALID_RANGE_MODE, "data: %s", s.data)
		}
		rs.SetMin(list[0], leader == TOK_RANGE_EX_L)
		rs.SetMax(nil, false)
		rs.SetLengthMode(mode)
	case 2:
		if IsNil(list[0]) && IsNil(list[1]) {
			return nil, nil
		}
		rs.SetMin(list[0], leader == TOK_RANGE_EX_L)
		rs.SetMax(list[1], tailer == TOK_RANGE_EX_R)
	default:
		return nil, codex.Errorf(ERROR__INVALID_RANGE_COUNT, "range list: %d", len(list))
	}
	return
}

func (s *_scanner) values() ([]*Literal, error) {
	if tok := s.Next(); tok != TOK_VALUES_L {
		return nil, codex.Errorf(
			ERROR__INVALID_VALUE_LEADER,
			"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
		)
	}

	var (
		values []*Literal
		value  *Literal
	)

LOOP:
	for tok := s.Peek(); ; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, codex.Errorf(
				ERROR__INVALID_VALUE_TAILER,
				"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
			)
		}
		switch tok {
		case ' ':
			s.Next()
		case TOK_SPLITTER:
			s.Next()
			values = append(values, value)
			value = nil
		case TOK_VALUES_R:
			s.Next()
			values = append(values, value)
			break LOOP
		default:
			text, err := s.literal()
			if err != nil {
				return nil, err
			}
			if value == nil {
				value = NewLiteral(text)
				continue
			}
			value.Append([]byte(text)...)
		}
	}

	return values, nil
}

func (s *_scanner) pattern() (Matcher, error) {
	if tok := s.Next(); tok != TOK_REGEXP {
		return nil, codex.Errorf(
			ERROR__INVALID_REGEXP_LEADER,
			"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
		)
	}

	b := &bytes.Buffer{}
	for tok := s.Peek(); tok != TOK_REGEXP; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, codex.Errorf(
				ERROR__INVALID_REGEXP_TAILER,
				"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
			)
		}
		if tok == '\\' {
			tok = s.Next()
			next := s.Next()
			// \/ -> /
			if next != '/' {
				b.WriteRune(tok)
			}
			b.WriteRune(next)
			continue
		}
		b.WriteRune(tok)
		s.Next()
	}
	s.Next()

	m := NewMatcher()
	if err := m.SetPattern(b.String()); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *_scanner) requirements() (Requirements, error) {
	r := NewRequirements()
	b := bytes.NewBuffer(nil)

	switch tok := s.Next(); tok {
	case TOK_EQUAL:
		tok = s.Next()
		if tok == '\'' {
			for tok = s.Peek(); tok != '\''; tok = s.Peek() {
				if tok == scanner.EOF {
					return nil, codex.Errorf(
						ERROR__INVALID_QUOTED_DEFAULT_TAILER,
						"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
					)
				}
				if tok == '\\' {
					tok = s.Next()
					next := s.Next()
					// \' -> '
					if next != '\'' {
						b.WriteRune(tok)
					}
					b.WriteRune(next)
					continue
				}
				b.WriteRune(tok)
				s.Next()
			}
			s.Next()
		} else {
			if tok != scanner.EOF && tok != TOK_PARAM_R && tok != TOK_SPLITTER {
				b.WriteRune(tok)
				text, err := s.literal()
				if err != nil {
					return nil, err
				}
				b.WriteString(text)
			}
		}
		r.SetDefaults(NewLiteral(b.String()))
	case TOK_OPTIONAL:
		r.SetOptional(true)
	default:
		return nil, codex.Errorf(
			ERROR__INVALID_REGEXP_TAILER,
			"data: %s pos: %d got: %s", s.data, s.Pos().Offset, string(tok),
		)
	}

	return r, nil
}
