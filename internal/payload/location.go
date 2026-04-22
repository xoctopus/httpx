package payload

import "slices"

const (
	HEADER = "header"
	PATH   = "path"
	QUERY  = "query"
	COOKIE = "cookie"
	BODY   = "body"
)

var Locations = slices.Values([]string{HEADER, PATH, QUERY, COOKIE, BODY})
