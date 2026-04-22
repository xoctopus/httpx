package path

import (
	"fmt"
	"strings"
)

type Matcher interface {
	MatchTo(m ValuesModifier, name string) (string, bool)
}

func NewMatcher(ss Segments, prefix string) Matcher {
	n := len(ss)

	m := &matcher{
		prefix:   prefix,
		Segments: make(Segments, 0, n),
	}

	for i, s := range ss {
		m.Segments = append(m.Segments, s)

		// not last
		if i != n-1 {
			if np, ok := s.(NamedSegment); ok && np.Multiple() {
				return &composedMatcher{
					l: m,
					r: NewMatcher(ss[i+1:], np.ParamName()),
				}
			}
		}
	}

	return m
}

type matcher struct {
	Segments
	prefix string
}

func (pm *matcher) MatchTo(vm ValuesModifier, name string) (string, bool) {
	parts := SplitPath(name)

	segn := len(pm.Segments)
	if len(parts) < segn {
		return "", false
	}

	strictPrefix := pm.prefix == ""

	offset := 0

	defer func() {
		if pm.prefix != "" {
			vm.SetPathValue(pm.prefix, strings.Join(parts[0:offset], "/"))
		}
	}()

	segi := 0
	for idx, part := range parts {
		segi = idx - offset
		if segi >= segn {
			return "", false
		}

		s := pm.Segments[segi]

		if np, ok := s.(NamedSegment); ok {
			if np.Multiple() {
				remain := strings.Join(parts[idx:], "/")
				vm.SetPathValue(np.ParamName(), remain)
				return remain, true
			}
			vm.SetPathValue(np.ParamName(), part)
			continue
		}

		if s.ParamString() != part {
			if strictPrefix {
				return "", false
			}
			offset++
		}
	}

	// make sure seg all matched
	if segi != segn-1 {
		return "", false
	}

	return "", true
}

type composedMatcher struct {
	l Matcher
	r Matcher
}

func (c *composedMatcher) String() string {
	return fmt.Sprintf("%s => %s", c.l, c.r)
}

func (c *composedMatcher) MatchTo(m ValuesModifier, name string) (string, bool) {
	remain, ok := c.l.MatchTo(m, name)
	if !ok {
		return "", ok
	}
	return c.r.MatchTo(m, remain)
}
