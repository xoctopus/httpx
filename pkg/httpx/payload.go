package httpx

import (
	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/payload/metadata"
	"github.com/xoctopus/httpx/internal/payload/transformer"
)

type (
	Content          = content.Content
	ContentReader    = content.Reader
	ContentProvider  = content.Provider
	ContentDescriber = content.Describer

	Transformer         = transformer.Transformer
	TransformerProvider = transformer.Provider

	Metadata        = metadata.Metadata
	MetadataCarrier = metadata.Carrier
)

var (
	MergeMetadata = metadata.Merge
)
