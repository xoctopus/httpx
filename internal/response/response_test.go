package response_test

import (
	"context"
	"net/http"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/response"
)

type CanWrite struct{}

func (CanWrite) WriteResponse(_ context.Context, rw http.ResponseWriter, _ *http.Request) error {
	_, err := rw.Write([]byte("CanWrite"))
	return err
}

func TestResponse(t *testing.T) {
	r := response.New(new(100))

	Expect(t, r.Underlying(), Equal(new(100)))
}
