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

func ErrorsFrom(err error) iter.Seq[error] {
	hasLocationOrPosition := func(x error) bool {
		_, hasPosition := x.(HasPosition)
		_, hasLocation := x.(HasLocation)
		return hasPosition || hasLocation
	}

	return func(yield func(error) bool) {
		if err == nil {
			return
		}

		switch x := err.(type) {
		case Describer:
			yield(err)
		case HasPosition:
			yield(err)
		case interface{ Unwrap() error }:
			if hasLocationOrPosition(err) {
				yield(err)
			}
			ue := x.Unwrap()
			if ue == nil {
				return
			}
			for e := range ErrorsFrom(ue) {
				yield(e)
			}
		case interface{ Unwrap() []error }:
			if hasLocationOrPosition(err) {
				yield(err)
			}
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
