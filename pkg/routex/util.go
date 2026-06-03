package routex

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/juju/ansiterm"

	"github.com/xoctopus/httpx/internal/payload/path"
)

func PathPrefix(segments path.Segments) string {
	s := &strings.Builder{}
	n := len(segments)

	for i, seg := range segments {
		switch x := seg.(type) {
		case path.NamedSegment:
			s.WriteString("/")
			if x.Multiple() && i == (n-1) {
				s.WriteString(x.PathString())
			} else {
				s.WriteString("{")
				s.WriteString(x.ParamName())
				s.WriteString("}")
			}
		default:
			s.WriteString("/")
			s.WriteString(x.PathString())
		}
	}

	return cmp.Or(s.String(), "/")
}

var colors = map[string]ansiterm.Color{
	http.MethodHead:   ansiterm.Cyan,
	http.MethodGet:    ansiterm.Blue,
	http.MethodPost:   ansiterm.Green,
	http.MethodPut:    ansiterm.Yellow,
	http.MethodPatch:  ansiterm.Magenta,
	http.MethodDelete: ansiterm.Red,
}

func newFormatter(method string, writer ...io.Writer) *formatter {
	f := &formatter{
		foreground: ansiterm.Gray,
		writer:     os.Stdout,
	}
	if c, ok := colors[method]; ok {
		f.foreground = c
	}
	if len(writer) > 0 && writer[0] != nil {
		f.writer = writer[0]
	}
	return f
}

type formatter struct {
	foreground ansiterm.Color
	writer     io.Writer
}

func (f formatter) Printf(msg string, args ...any) (int, error) {
	if x, ok := f.writer.(interface{ SetForeground(ansiterm.Color) }); ok {
		x.SetForeground(f.foreground)
		defer func() { x.SetForeground(ansiterm.Default) }()
	}
	return fmt.Fprintf(f.writer, msg, args...)
}
