package httpx

import (
	"net"
	"net/http"
	"strings"

	"github.com/xoctopus/httpx/internal/client"
)

type (
	Client = client.Client
	Result = client.Result
)

// ClientIP
// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
func ClientIP(r *http.Request) string {
	ip := ClientIPByHeaderForwardedFor(r.Header.Get("X-Forwarded-For"))
	if ip != "" {
		return ip
	}

	ip = ClientIPByHeaderRealIP(r.Header.Get("X-Real-IP"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// ClientIPByHeaderForwardedFor
// X-Forwarded-For: client, proxy1, ..., proxyN
func ClientIPByHeaderForwardedFor(v string) string {
	if before, _, ok := strings.Cut(v, ","); ok {
		return before
	}
	return strings.TrimSpace(v)
}

// ClientIPByHeaderRealIP
// X-Real-IP: 203.0.113.195, 2001:db8:85a3:8d3:1319:8a2e:370:7348
func ClientIPByHeaderRealIP(v string) string {
	return strings.TrimSpace(v)
}
