package upgrader

import "net/http"

type Upgrader interface {
	Upgrade(http.ResponseWriter, *http.Request) error
}
