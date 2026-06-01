package status

import "iter"

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

// HasPosition presents error position field or key.
type HasPosition interface {
	Position() string
}

// HasLocation presents error location. see payload.Locations
type HasLocation interface {
	Location() string
}

type HasCodeText interface {
	// StatusText returns code text. eg: http.StatusOK => http.StatusText(http.StatusOK)
	StatusText() string
}

type CanUnmarshalResponse interface {
	UnmarshalResponse(code int, data []byte) error
}

func ErrorsFrom(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		switch x := err.(type) {
		case nil:
			return
		case HasPosition:
			yield(err)
		case Describer:
			yield(err)
		case interface{ Unwrap() error }:
			if _, ok := err.(HasLocation); ok {
				yield(err)
			}
			if ue := x.Unwrap(); ue != nil {
				for e := range ErrorsFrom(ue) {
					yield(e)
				}
			}
		case interface{ Unwrap() []error }:
			for _, ue := range x.Unwrap() {
				if ue != nil {
					for e := range ErrorsFrom(ue) {
						yield(e)
					}
				}
			}
		default:
			yield(err)
		}
	}
}
