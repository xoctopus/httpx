package content

import (
	"context"
	"io"
	"net/http"
	"reflect"
)

type (
	// Content payload content
	Content interface {
		// Type presents content-type value. eg: json, text, etc.
		Type() string
		// Length presents content length limitation. -1 is unlimited
		Length() int64

		io.ReadCloser
	}

	Builder interface {
		Content

		SetContentType(string)
		SetLength(int64)
		SetReadCloser(io.ReadCloser)
	}

	Modifier interface {
		SetContentType(string)
	}

	Applier interface {
		ApplyHeader(header http.Header)
	}

	Describer interface {
		ContentType() string
	}

	Reader interface {
		Into(context.Context, io.ReadCloser, any) error
	}

	Writer interface {
		From(context.Context, io.WriteCloser, any) error
	}

	Provider interface {
		Prepare(ctx context.Context, src any) (Content, error)
	}
)

// WithFilename for getting multipart content filename
type WithFilename interface {
	Filename() string
}

// FilenameModifier for setting multipart content filename
type FilenameModifier interface {
	SetFilename(string)
}

type WithHeader interface {
	Header() http.Header
}

const (
	MaxBufferSize = 32 << 20 // 32 MB
	MaxFilesize   = 32 << 20 // 32 MB
)

var (
	TWithFilename = reflect.TypeFor[WithFilename]()
	TWithHeader   = reflect.TypeFor[WithHeader]()
)
