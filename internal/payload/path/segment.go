package path

import (
	"fmt"
	"strings"
)

type Segment interface {
	ParamString() string
}

func NewSegment(part string) Segment {
	if len(part) == 0 {
		return nil
	}

	if part[0] == '{' {
		np := &namedSegment{name: part[1 : len(part)-1]}
		if strings.HasSuffix(np.name, "...") {
			np.name = np.name[0 : len(np.name)-3]
			np.multiple = true
		}
		return np
	}

	return segment(part)

}

type segment string

func (s segment) ParamString() string {
	return string(s)
}

type NamedSegment interface {
	Segment
	ParamName() string
	Multiple() bool
}

type namedSegment struct {
	name     string
	multiple bool
}

func (np namedSegment) Multiple() bool {
	return np.multiple
}

func (np namedSegment) ParamName() string {
	return np.name
}

func (np namedSegment) ParamString() string {
	s := "{%s}"
	if np.multiple {
		s = "{%s...}"
	}
	return fmt.Sprintf(s, np.name)
}

func ParseSegments(p string) Segments {
	parts := SplitPath(p)

	ss := make(Segments, len(parts))

	for i, part := range parts {
		if np := NewSegment(part); np != nil {
			ss[i] = np
		}
	}

	return ss
}

type Segments []Segment

func (ss Segments) PathValues(resource string) (Values, error) {
	params := Values{}

	_, ok := ss.MatchTo(params, resource)
	if !ok {
		return nil, fmt.Errorf("pathname %s is not match %s", resource, ss.ParamString())
	}

	return params, nil
}

func (ss Segments) MatchTo(vm ValuesModifier, res string) (string, bool) {
	return NewMatcher(ss, "").MatchTo(vm, res)
}

func (ss Segments) ParamString() string {
	b := &strings.Builder{}

	b.WriteString("/")
	for i, s := range ss {
		if i > 0 {
			b.WriteString("/")
		}
		b.WriteString(s.ParamString())
	}
	return b.String()
}

func (ss Segments) Encode(vs Values) string {
	x := make(Segments, len(ss))

	for idx, s := range ss {
		if named, ok := s.(NamedSegment); ok {
			v := vs[named.ParamName()]
			if v == "" {
				v = "-"
			}
			x[idx] = segment(v)
			continue
		}
		x[idx] = s
	}

	return x.ParamString()
}
