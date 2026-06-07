package path

import (
	"cmp"
	"fmt"
	"iter"
	"strings"
)

type Segment interface {
	PathString() string
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

func (s segment) PathString() string {
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

func (np namedSegment) PathString() string {
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

func (ss Segments) Chunk() iter.Seq[Segments] {
	return func(yield func(Segments) bool) {
		lastOmit := 0

		for i, s := range ss {
			if named, ok := s.(NamedSegment); ok {
				if named.Multiple() {
					if !yield(ss[lastOmit : i+1]) {
						return
					}
					lastOmit = i + 1
				}
			}
		}

		if lastOmit > 0 {
			if lastOmit < len(ss) {
				if !yield(ss[lastOmit:]) {
					return
				}
			}
		} else {
			if !yield(ss[:]) {
				return
			}
		}
	}
}

func (ss Segments) PathValues(resource string) (Values, error) {
	params := Values{}

	_, ok := ss.MatchTo(params, resource)
	if !ok {
		return nil, fmt.Errorf("pathname %s is not match %s", resource, ss.PathString())
	}

	return params, nil
}

func (ss Segments) MatchTo(vm ValuesModifier, res string) (string, bool) {
	return NewMatcher(ss, "").MatchTo(vm, res)
}

func (ss Segments) PathString() string {
	b := &strings.Builder{}

	b.WriteString("/")
	for i, s := range ss {
		if i > 0 {
			b.WriteString("/")
		}
		b.WriteString(s.PathString())
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

	return x.PathString()
}

func Prefix(segments Segments) string {
	s := &strings.Builder{}
	n := len(segments)

	for i, seg := range segments {
		switch x := seg.(type) {
		case NamedSegment:
			s.WriteString("/")
			if x.Multiple() && i == (n-1) {
				s.WriteString(x.PathString())
			} else {
				s.WriteString("{")
				s.WriteString(x.ParamName())
				s.WriteString("}")
			}
		default:
			s.WriteString("/")
			s.WriteString(x.PathString())
		}
	}

	return cmp.Or(s.String(), "/")
}
