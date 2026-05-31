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
		ContentType() string
		ContentLength() int64

		io.ReadCloser
	}

	Builder interface {
		Content

		SetContentType(string)
		SetContentLength(int64)
		SetReadCloser(io.ReadCloser)
	}

	Modifier interface {
		SetContentType(string)
	}

	HeaderApplier interface {
		ApplyHeader(header http.Header)
	}

	MediaTypeDescriber interface {
		ContentType() string
	}

	MediaTypeModifier interface {
		SetContentType(string)
	}

	LengthDescriber interface {
		ContentLength() int64
	}

	LengthModifier interface {
		SetContentLength(int64)
	}

	// Reader defined the behavior of reading data from an io.ReadCloser and
	// deserializing into `dst`.
	Reader interface {
		Into(ctx context.Context, rc io.ReadCloser, dst any) error
	}

	ReaderFrom interface {
		ReadFrom(rc io.ReadCloser) (int64, error)
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
