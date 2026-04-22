package transformer

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"
	"github.com/xoctopus/x/syncx"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/payload/content"
)

type Provider interface {
	Names() []string
	Transformer() (Transformer, error)
}

type Transformer interface {
	Media() string

	content.Reader
	content.Provider
}

func Register(p Provider) {
	for _, name := range p.Names() {
		gTransformers.providers[name] = p
	}
}

func New(rtyp reflect.Type, typ, action string) (Transformer, error) {
	return gTransformers.New(Option{
		action: action,
		rtyp:   reflectx.DerefPointer(rtyp),
		typ:    typ,
	})
}

type transformers struct {
	providers map[string]Provider
	instances syncx.Map[Option, func() (Transformer, error)]
}

var gTransformers = &transformers{
	providers: make(map[string]Provider),
	instances: syncx.NewXmap[Option, func() (Transformer, error)](),
}

func (ts *transformers) New(i Option) (Transformer, error) {
	newer := func() (Transformer, error) {
		media := i.media()
		p, ok := ts.providers[media]
		if !ok {
			return nil, fmt.Errorf("unknown media type %s to %s", media, i.typ)
		}
		return p.Transformer()
	}

	f, _ := ts.instances.LoadOrStore(i, sync.OnceValues(newer))
	return f()
}

const (
	ForUnmarshalling = "unmarshal"
	ForMarshalling   = "marshal"
)

type Option struct {
	action string
	rtyp   reflect.Type
	typ    string
}

func (i *Option) typImpl(dst reflect.Type) bool {
	return i.rtyp.Implements(dst)
}

func (i *Option) ptrImpl(dst reflect.Type) bool {
	return reflect.PointerTo(i.rtyp).Implements(dst)
}

func (i *Option) impl(dst reflect.Type) bool {
	return i.typImpl(dst) || i.ptrImpl(dst)
}

func (i *Option) media() string {
	media := i.typ
	if strings.HasSuffix(media, "+json") {
		media = "json"
	}

	if media == "" {
		switch i.rtyp {
		case tBytes:
			media = "octet"
		case tString:
			media = "plain"
		}
	}

	if media == "" {
		switch i.action {
		case ForMarshalling:
			if i.impl(tReadCloser) {
				media = "octet"
			} else if i.impl(tTextMarshaler) {
				media = "plain"
			}
		default:
			must.BeTrue(i.action == ForUnmarshalling)
			if i.impl(tReadCloser) {
				media = "octet"
			} else if i.ptrImpl(tTextUnmarshaler) {
				media = "plain"
			} else if i.ptrImpl(tJSONUnmarshaler) {
				media = "json"
			}
		}
	}

	if media == "" {
		switch i.rtyp.Kind() {
		case reflect.String:
			media = "plain"
		default:
			media = "json"
		}
	}
	return media
}

var (
	tString              = reflect.TypeFor[string]()
	tBytes               = reflect.TypeFor[[]byte]()
	tReadCloser          = reflect.TypeFor[io.ReadCloser]()
	tTextMarshaler       = reflect.TypeFor[encoding.TextMarshaler]()
	tTextUnmarshaler     = reflect.TypeFor[encoding.TextUnmarshaler]()
	tJSONUnmarshaler     = reflect.TypeFor[json.Unmarshaler]()
	tJSONUnmarshalerFrom = reflect.TypeFor[json.UnmarshalerFrom]()
)
