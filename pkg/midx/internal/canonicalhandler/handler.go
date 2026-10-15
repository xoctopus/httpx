package canonicalhandler

import (
	"net/http"
	"net/url"
	"strings"
)

type Canonical struct {
	Handler http.Handler
	Domain  string
	Code    int
}

func (c Canonical) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dest, err := url.Parse(c.Domain)
	if err != nil {
		// Call the next handler if the provided domain fails to parse.
		c.Handler.ServeHTTP(w, r)
		return
	}

	if dest.Scheme == "" || dest.Host == "" {
		// Call the next handler if the scheme or host are empty.
		// Note that url.Parse won't fail on in this case.
		c.Handler.ServeHTTP(w, r)
		return
	}

	if !strings.EqualFold(cleanHost(r.Host), dest.Host) {
		// Re-build the destination URL
		u := dest.Scheme + "://" + dest.Host + r.URL.Path
		if r.URL.RawQuery != "" {
			u += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, u, c.Code)
		return
	}

	c.Handler.ServeHTTP(w, r)
}

// cleanHost cleans invalid Host headers by stripping anything after '/' or ' '.
// This is backported from Go 1.5 (in response to issue #11206) and attempts to
// mitigate malformed Host headers that do not match the format in RFC7230.
func cleanHost(in string) string {
	if i := strings.IndexAny(in, " /"); i != -1 {
		return in[:i]
	}
	return in
}
