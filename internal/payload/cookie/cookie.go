package cookie

import "net/http"

type Describer interface {
	Cookies() []*http.Cookie
}

type Modifier interface {
	SetCookies([]*http.Cookie)
}
