package client

import (
	"context"

	"github.com/xoctopus/httpx/internal/payload/metadata"
)

type Client interface {
	Do(ctx context.Context, req any, metas ...metadata.Metadata) Result
}

type Result interface {
	Into(v any) (metadata.Metadata, error)
}
