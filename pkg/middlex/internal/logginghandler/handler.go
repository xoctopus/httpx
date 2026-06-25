package logginghandler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/xoctopus/logx"
)

type ResponseWriter interface {
	http.ResponseWriter

	// StatusCode
	StatusCode() int
	// StatusError if status code >= 400 treat write content as errorString
	StatusError() error
}

func NewResponseWriter(rw http.ResponseWriter) ResponseWriter {
	h, hok := rw.(http.Hijacker)
	if !hok {
		h = nil
	}

	f, fok := rw.(http.Flusher)
	if !fok {
		f = nil
	}

	return &responseWriter{
		ResponseWriter: rw,
		Hijacker:       h,
		Flusher:        f,
	}
}

type responseWriter struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher

	headerWritten bool
	status        int
	written       int64
	err           error
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.headerWritten {
		rw.ResponseWriter.WriteHeader(code)
		rw.status = code
		rw.headerWritten = true
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.status == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	if rw.err == nil && rw.status >= http.StatusBadRequest {
		rw.err = errors.New(strings.TrimSpace(string(data)))
	}
	n, err := rw.ResponseWriter.Write(data)
	rw.written += int64(n)
	return n, err
}

func (rw *responseWriter) StatusCode() int {
	return rw.status
}

func (rw *responseWriter) StatusError() error {
	return rw.err
}

func OmitAuthorization(u *url.URL) string {
	q := u.Query()

	q.Del("authorization")
	q.Del("x-param-header-Authorization")

	u.RawQuery = q.Encode()
	return u.String()
}

var LogLevels = map[string]logx.LogLevel{
	"info":    logx.LogLevelInfo,
	"inf":     logx.LogLevelInfo,
	"debug":   logx.LogLevelDebug,
	"deb":     logx.LogLevelDebug,
	"warn":    logx.LogLevelWarn,
	"warning": logx.LogLevelWarn,
	"wrn":     logx.LogLevelWarn,
	"error":   logx.LogLevelError,
	"err":     logx.LogLevelError,
	"fatal":   logx.LogLevelError,
	"panic":   logx.LogLevelError,
}
