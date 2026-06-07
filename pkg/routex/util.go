package routex

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/juju/ansiterm"
)

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
