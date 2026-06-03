package path

import (
	"strings"
)

type Matcher interface {
	MatchTo(m ValuesModifier, uri string) (string, bool)
}

func NewMatcher(ss Segments, prefix string) Matcher {
	n := len(ss)

	m := &matcher{
		prefix:   prefix,
		Segments: make(Segments, 0, n),
	}

	for i, s := range ss {
		m.Segments = append(m.Segments, s)

		// net/http/pattern.go named multiple pattern should be the last one?
		if i != n-1 {
			if np, ok := s.(NamedSegment); ok && np.Multiple() {
				return &composedMatcher{
					head: m,
					tail: NewMatcher(ss[i+1:], np.ParamName()),
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

func (pm *matcher) MatchTo(vm ValuesModifier, uri string) (string, bool) {
	parts := SplitPath(uri)

	segn := len(pm.Segments)
	if len(parts) < segn {
		return "", false
	}

	strict := pm.prefix == ""

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

		if s.PathString() != part {
			if strict {
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
	head Matcher
	tail Matcher
}

func (c *composedMatcher) MatchTo(m ValuesModifier, uri string) (string, bool) {
	remain, ok := c.head.MatchTo(m, uri)
	if !ok {
		return "", ok
	}
	return c.tail.MatchTo(m, remain)
}
