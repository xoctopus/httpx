package status

import (
	"encoding/json/jsontext"
	"net/http"

	"github.com/xoctopus/x/codex"
)

type Describer interface {
	StatusCode() int
}

type Status interface {
	Describer
	HasCodeText

	IsValid() bool
}

type Error interface {
	Describer
	HasCodeText
	error
}

type Modifier interface {
	SetStatusCode(int)
}

type DetailSetter interface {
	SetDetail(detail string)
}

type TitleSetter interface {
	SetTitle(title string)
}

// HasPointer presents error position by a JSON Pointer (RFC 6901) in the
// request document, see Detail.Pointer
type HasPointer interface {
	Pointer() jsontext.Pointer
}

type HasCodeText interface {
	// StatusText returns code text. eg: http.StatusOK => http.StatusText(http.StatusOK)
	StatusText() string
}

type CanUnmarshalResponse interface {
	UnmarshalResponse(code int, data []byte) error
}

// resolution is the state accumulated by walking an error chain from outside in.
// status and the identity layer are resolved independently: each takes the first
// node providing it and is never overwritten by inner nodes.
type resolution struct {
	status int

	ident  bool // the identity layer (the outermost codex.Error) is found
	code   int64
	key    string
	title  string
	detail string

	pointer jsontext.Pointer
}

func (r *resolution) visit(err error) {
	if r.status <= 0 {
		if x, ok := err.(Describer); ok {
			r.status = x.StatusCode()
		}
	}

	if x, ok := err.(codex.Error); ok {
		if r.status <= 0 {
			var d Describer
			if x.As(&d) {
				r.status = d.StatusCode()
			}
		}
		if !r.ident {
			r.ident = true
			r.code = x.Code()
			r.detail = x.Detail()
			var id codex.HasIdentifierCode
			if x.As(&id) {
				r.key = id.CodeIdentifier()
			}
			var hint codex.HasHintCode
			if x.As(&hint) {
				r.title = hint.CodeHint()
			}
		}
	}

	if len(r.pointer) == 0 {
		if x, ok := err.(HasPointer); ok {
			r.pointer = x.Pointer()
		}
	}
}

// statusCode returns the resolved status, 400 for errors located in the
// request, otherwise 500.
func (r *resolution) statusCode() int {
	switch {
	case r.status > 0:
		return r.status
	case len(r.pointer) > 0:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (r *resolution) asDetail(desc, src string) *Detail {
	return &Detail{
		Code:    r.code,
		Key:     r.key,
		Title:   r.title,
		Detail:  r.detail,
		Pointer: r.pointer,
		Src:     src,
		Desc:    desc,
		Status:  r.statusCode(),
	}
}

// resolve walks err from outside in. it returns the resolution of the chain
// above the first errors.Join (or of the whole chain without Join), and one
// Detail per branch. branches inherit the status of the outer chain but resolve
// their own identity layer. a downstream *Detail is kept as an entry as is.
func resolve(err error, r resolution, src string) (resolution, []*Detail) {
	desc := err.Error()
	for e := err; e != nil; {
		if d, ok := e.(*Detail); ok {
			return r, []*Detail{d}
		}
		r.visit(e)

		switch x := e.(type) {
		case interface{ Unwrap() []error }:
			var entries []*Detail
			for _, branch := range x.Unwrap() {
				if branch != nil {
					_, ds := resolve(branch, resolution{status: r.status, pointer: r.pointer}, src)
					entries = append(entries, ds...)
				}
			}
			return r, entries
		case interface{ Unwrap() error }:
			e = x.Unwrap()
		default:
			e = nil
		}
	}
	return r, []*Detail{r.asDetail(desc, src)}
}
