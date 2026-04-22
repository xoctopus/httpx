package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

func ConvertTransportForH2C(t http.RoundTripper) http.RoundTripper {
	if x, ok := t.(*http.Transport); ok {
		if !x.DisableKeepAlives {
			return &http2.Transport{
				DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
					return x.DialContext(ctx, network, addr)
				},
				IdleConnTimeout: x.IdleConnTimeout,
				ReadIdleTimeout: x.ResponseHeaderTimeout,
				AllowHTTP:       true,
			}
		}
	}
	return t
}

func H2CTransport(t http.RoundTripper) http.RoundTripper {
	if x, ok := t.(*http.Transport); ok {
		if !x.DisableKeepAlives {
			return &http2.Transport{
				DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
					return x.DialContext(ctx, network, addr)
				},
				IdleConnTimeout: x.IdleConnTimeout,
				ReadIdleTimeout: x.ResponseHeaderTimeout,
				AllowHTTP:       true,
			}
		}
	}
	return t
}
