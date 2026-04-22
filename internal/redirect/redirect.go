package redirect

import (
	"net/url"

	"github.com/xoctopus/httpx/internal/status"
)

type Describer interface {
	status.Describer

	Location() *url.URL
}

type Modifier interface {
	SetLocation(*url.URL)
}
