package middlex

import (
	"compress/gzip"
	"net/http"

	"github.com/xoctopus/httpx/pkg/middlex/internal/compresshandler"
)

func Compress(level int) func(h http.Handler) http.Handler {
	return compresshandler.HandlerLevel(level)
}

func DefaultCompress() func(h http.Handler) http.Handler {
	return compresshandler.HandlerLevel(gzip.DefaultCompression)
}
