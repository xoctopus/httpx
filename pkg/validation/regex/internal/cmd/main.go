package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/stringsx"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "regexp.go", nil, parser.ParseComments)
	must.NoError(err)

	b := bytes.NewBuffer(nil)

	defines := make([]string, 0)
	for name, obj := range f.Scope.Objects {
		if obj.Kind == ast.Con {
			defines = append(defines, name)
		}
	}
	slices.Sort(defines)

	b.WriteString("package regex\n\n")
	b.WriteString("import (\n")
	b.WriteString(`	"github.com/xoctopus/httpx/internal/validation"` + "\n")
	b.WriteString(`	"github.com/xoctopus/httpx/pkg/validation/validators"` + "\n")
	b.WriteString(")\n\n")

	for _, name := range defines {
		if !strings.HasPrefix(name, "regex") {
			continue
		}
		key := strings.TrimPrefix(name, "regex")      // Alpha
		providerName := key + "RegexProvider"         // AlphaRegexProvider
		validatorName := stringsx.LowerCamelCase(key) // alpha

		b.WriteString(fmt.Sprintf("func init() { validation.Register(%s) }\n\n", providerName))
		b.WriteString(fmt.Sprintf("var %s = validators.NewRegexpProvider(%s, %q, %q)\n\n", providerName, name, validatorName, ""))
	}

	file := must.NoErrorV(os.OpenFile("regexp_genx.go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666))
	defer func() { _ = file.Close() }()

	must.NoErrorV(io.Copy(file, b))
}
