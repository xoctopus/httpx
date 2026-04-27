package httpx

import (
	"net/http"

	"github.com/xoctopus/httpx/internal/response"
)

type (
	Response[T any] = response.Response[T]
	ResponseApplier = response.Applier
)

func WithStatusCode(statusCode int) ResponseApplier {
	return func(m response.Modifier) {
		m.SetStatusCode(statusCode)
	}
}

func WithCookies(cookies ...*http.Cookie) ResponseApplier {
	return func(m response.Modifier) {
		m.SetCookies(cookies)
	}
}

func WithContentType(contentType string) ResponseApplier {
	return func(m response.Modifier) {
		m.SetContentType(contentType)
	}
}

func WithMetadata(key string, values ...string) ResponseApplier {
	return func(m response.Modifier) {
		m.SetMetadata(key, values...)
	}
}

func WrapResponse[T any](v T, appliers ...ResponseApplier) Response[T] {
	rsp := response.New[T](v)
	mod := rsp.(response.Modifier)

	for _, apply := range appliers {
		apply(mod)
	}

	return rsp
}

func ErrorResponse(err error, appliers ...ResponseApplier) Response[error] {
	rsp := response.New[error](err)
	mod := rsp.(response.Modifier)

	for _, apply := range appliers {
		apply(mod)
	}

	// // TODO
	// if rsp.StatusCode() == 0 {
	// }

	return rsp
}
