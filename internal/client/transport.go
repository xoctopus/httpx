package client

import "net/http"

func H2CTransport(t http.RoundTripper) http.RoundTripper {
	x, ok := t.(*http.Transport)
	if !ok || x.DisableKeepAlives {
		return t
	}

	tr := x.Clone()
	p := &http.Protocols{}
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)
	tr.Protocols = p

	if tr.ResponseHeaderTimeout > 0 {
		if tr.HTTP2 == nil {
			tr.HTTP2 = &http.HTTP2Config{}
		}
		tr.HTTP2.SendPingTimeout = tr.ResponseHeaderTimeout
	}

	return tr
}
