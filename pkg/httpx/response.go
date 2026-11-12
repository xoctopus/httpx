package httpx

import (
	"fmt"
	"net/http"

	"github.com/xoctopus/httpx/internal/response"
)

type (
	Response[T any] = response.Response[T]
	ResponseApplier = response.Applier
)

func WithResponseStatusCode[Code ~int](code Code) ResponseApplier {
	return func(m response.Modifier) {
		m.SetStatusCode(int(code))
	}
}

func WithResponseCookies(cookies ...*http.Cookie) ResponseApplier {
	return func(m response.Modifier) {
		m.SetCookies(cookies)
	}
}

func WithResponseContentType(t string) ResponseApplier {
	return func(m response.Modifier) {
		m.SetContentType(t)
	}
}

func WithResponseTitle(title string) ResponseApplier {
	return func(m response.Modifier) {
		m.SetTitle(title)
	}
}

func WithResponseDetail(format string, args ...any) ResponseApplier {
	return func(m response.Modifier) {
		m.SetDetail(fmt.Sprintf(format, args...))
	}
}

func WithResponseMetadata(key string, values ...string) ResponseApplier {
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

	return rsp
}
