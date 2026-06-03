package jsonschema

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// URI uri
type URI url.URL

func (u *URI) UnmarshalText(b []byte) error {
	x, err := url.Parse(string(b))
	if err != nil {
		return err
	}
	*u = URI(*x)
	return nil
}

func (u URI) MarshalText() ([]byte, error) {
	x := url.URL(u)
	return []byte(x.String()), nil
}

func ParseURIRef(u string) (*URIRef, error) {
	x, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return (*URIRef)(x), nil
}

// URIRef uri-reference
type URIRef url.URL

func (u *URIRef) RefName() string {
	if u.Fragment == "" {
		return ""
	}
	// last part
	parts := strings.Split(u.Fragment, "/")
	return parts[len(parts)-1]
}

func (u *URIRef) UnmarshalText(b []byte) error {
	x, err := url.Parse(string(b))
	if err != nil {
		return err
	}
	*u = URIRef(*x)
	return nil
}

func (u URIRef) MarshalText() ([]byte, error) {
	x := url.URL(u)

	if x.Scheme == "" {
		var buf strings.Builder

		if x.Path != "" {
			buf.WriteString(x.Path)
		}

		if x.RawQuery != "" {
			buf.WriteString("?")
			buf.WriteString(x.RawQuery)
		}

		if x.Fragment != "" {
			buf.WriteString("#")
			buf.WriteString(x.Fragment)
		}

		return []byte(buf.String()), nil
	}

	return []byte(x.String()), nil
}

// Anchor openapi:anchor
type Anchor string

var (
	anchorPattern      = "^[A-Za-z_][-A-Za-z0-9._]*$"
	regexAnchorPattern = regexp.MustCompile(anchorPattern)
)

func (s *Anchor) UnmarshalText(text []byte) error {
	if regexAnchorPattern.Match(text) {
		*s = Anchor(text)
	}
	return errors.New("invalid anchor string")
}

func (Anchor) OpenAPISchema() Schema {
	return &StringType{
		Type:    "string",
		Pattern: anchorPattern,
	}
}
