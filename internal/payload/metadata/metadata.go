package metadata

import "net/url"

type Carrier interface {
	Meta() Metadata
}

type Modifier interface {
	SetMetadata(k string, vs ...string)
}

func Merge(ms ...Metadata) Metadata {
	m := Metadata{}
	for _, _m := range ms {
		m.Merge(_m)
	}
	return m
}

type Metadata map[string][]string

func (m Metadata) String() string {
	return url.Values(m).Encode()
}

func (m Metadata) Del(k string) { delete(m, k) }

func (m Metadata) Add(k, v string) {
	if vs, ok := m[k]; ok {
		m[k] = append(vs, v)
	} else {
		m[k] = []string{v}
	}
}

func (m Metadata) Set(k string, vs ...string) {
	m[k] = vs
}

func (m Metadata) Has(k string) bool {
	_, has := m[k]
	return has
}

func (m Metadata) Get(k string) string {
	if vs, ok := m[k]; ok && len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (m Metadata) Merge(m2 Metadata) {
	for k, vs := range m2 {
		m.Set(k, vs...)
	}
}
