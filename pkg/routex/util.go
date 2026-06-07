package routex

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
)

var formatters = map[string]*color.Color{
	http.MethodHead:   color.New( /*color.Bold,*/ color.FgCyan),
	http.MethodGet:    color.New( /*color.Bold,*/ color.FgBlue),
	http.MethodPost:   color.New( /*color.Bold,*/ color.FgGreen),
	http.MethodPut:    color.New( /*color.Bold,*/ color.FgYellow),
	http.MethodPatch:  color.New( /*color.Bold,*/ color.FgMagenta),
	http.MethodDelete: color.New( /*color.Bold,*/ color.FgRed),
	"":                color.New( /*color.Bold,*/ color.FgWhite),
}

type meta struct {
	Method    string
	Path      string
	Summary   string
	Operators string

	method string
}

type printer struct {
	meta []*meta
}

func (p *printer) Add(m *meta) {
	p.meta = append(p.meta, m)
}

func (p *printer) Flush() {
	var (
		maxMethod  = 0
		maxPath    = 0
		maxSummary = 0
		padding    = 2
	)

	for _, m := range p.meta {
		m.method = m.Method
		if n := len(m.Method); n > maxMethod {
			maxMethod = n
		}
		if n := len(m.Path); n > maxPath {
			maxPath = n
		}
		l := len(m.Summary)
		n := len([]rune(m.Summary))
		if l != n {
			n = n * 2
		}
		if n > maxSummary {
			maxSummary = n
		}
	}

	for _, m := range p.meta {
		m.Method = fmt.Sprintf("%-*s", maxMethod+padding, m.Method)
		m.Path = fmt.Sprintf("%-*s", maxPath+padding, m.Path)

		l := len(m.Summary)
		n := len([]rune(m.Summary))
		if l != n {
			n = n * 2 // Chinese character. Final width
		}
		c := maxSummary - n + padding

		// s := MethodColorFormatter(m.method)("%s%s", m.Method, m.Path)
		f := formatters[m.method]
		if f == nil {
			f = formatters[""]
		}
		space := strings.Repeat(" ", c)
		s := f.Sprintf("%s%s%s%s", m.Method, m.Path, m.Summary, space)
		fmt.Printf("%s%s\n", s, m.Operators)
	}
	fmt.Print("\n")
}
