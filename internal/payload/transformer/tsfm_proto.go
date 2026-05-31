package transformer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"

	"google.golang.org/protobuf/proto"

	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/scanner"
)

func init() {
	Register(&protoP{})
}

type protoP struct{}

func (p *protoP) Names() []string {
	return []string{
		"application/x-protobuf",
		"protobuf",
		"x-protobuf",
		"proto",
	}
}

func (p *protoP) Transformer() (Transformer, error) {
	return &protoT{
		media: p.Names()[0],
	}, nil
}

type protoT struct {
	media string
}

func (p *protoT) Media() string {
	return p.media
}

func (p *protoT) Into(_ context.Context, r io.ReadCloser, v any) error {
	defer func() { _ = r.Close() }()

	u := scanner.UnwrapField(v)
	rv, ok := u.(reflect.Value)
	if ok {
		u = rv.Interface()
	}

	m, ok := u.(proto.Message)
	if !ok {
		return fmt.Errorf("expect a proto.Message value but got %T", u)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return proto.Unmarshal(data, m)
}

func (p *protoT) Prepare(_ context.Context, v any) (content.Content, error) {
	b := content.NewBuilder(p.media)

	u := scanner.UnwrapField(v)
	if rv, ok := u.(reflect.Value); ok {
		u = rv.Addr().Interface()
	}
	m, ok := u.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("expect a proto.Message value but got %T", u)
	}
	data, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	b.SetContentLength(int64(len(data)))
	b.SetReadCloser(io.NopCloser(bytes.NewBuffer(data)))
	return b, nil
}
