package va

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation/rule"
)

func NewFloatScaleVa(r rule.Rule) (*FloatScaleVa, error) {
	fs := &FloatScaleVa{}
	fs.bits = 32
	if strings.HasSuffix(r.Name(), "64") || r.Name() == "double" {
		fs.bits = 64
	}

	params := r.Parameters()
	if len(params) == 0 {
		return fs, nil
	}

	if len(params) != 1 && len(params) != 2 {
		return nil, codex.Errorf(ERROR__INVALID_FLOAT_SCALE, "got %d parameters", len(params))
	}

	if len(params) >= 1 && !rule.IsNil(params[0]) {
		lit, ok := params[0].(*rule.Literal)
		if !ok {
			return nil, codex.Errorf(ERROR__INVALID_FLOAT_SCALE, "invalid parameter got %s", params[0].Type())
		}

		data := lit.String()
		v, err := strconv.ParseUint(data, 10, 6)
		if err != nil {
			return nil, codex.Wrapf(ERROR__INVALID_FLOAT_SCALE, err, "too much digital max 63 got: %s", data)
		}
		fs.digital = new(uint(v))
	}

	if len(params) == 2 && !rule.IsNil(params[1]) {
		lit, ok := params[1].(*rule.Literal)
		if !ok {
			return nil, codex.Errorf(ERROR__INVALID_FLOAT_SCALE, "invalid parameter got %s", params[1].Type())
		}

		data := lit.String()
		v, err := strconv.ParseUint(data, 10, 4)
		if err != nil {
			return nil, codex.Wrapf(ERROR__INVALID_FLOAT_SCALE, err, "too much decimal max 31 got: %s", data)
		}
		fs.decimel = new(uint(v))
		if fs.digital != nil && *fs.decimel >= *fs.digital {
			return nil, codex.Errorf(
				ERROR__INVALID_FLOAT_SCALE,
				"got digital: %d, decimal: %d", *fs.digital, *fs.decimel,
			)
		}
	}
	return fs, nil
}

type FloatScaleVa struct {
	bits    int
	digital *uint
	decimel *uint
}

func (fs *FloatScaleVa) BuildTo(b rule.Builder) {
	if fs != nil {
		if fs.digital != nil {
			b.AddParameters(rule.NewLiteral(strconv.Itoa(int(*fs.digital))))
			if fs.decimel != nil {
				b.AddParameters(rule.NewLiteral(strconv.Itoa(int(*fs.decimel))))
			}
		}
	}
}

func ExtractFloatScales(v string) (w, p uint) {
	v = strings.TrimPrefix(v, "-")

	dot := strings.IndexByte(v, '.')
	if dot == -1 {
		v = strings.TrimLeft(v, "0")
		if len(v) == 0 {
			return 1, 0
		}
		return uint(len(v)), 0
	}

	iw, fw := len(strings.TrimLeft(v[:dot], "0")), len(v[dot+1:])
	if iw == 0 {
		iw = 1
	}
	return uint(iw + fw), uint(fw)
}

func (fs *FloatScaleVa) Validate(v string) error {
	_, err := strconv.ParseFloat(v, fs.bits)
	if err != nil {
		return codex.Wrapf(ERROR__OUT_OF_FLOAT_SCALE, err, "float bits: %d", fs.bits)
	}
	w, p := ExtractFloatScales(v)
	if fs.digital != nil && w > *fs.digital ||
		fs.decimel != nil && p > *fs.decimel {
		return codex.Errorf(ERROR__OUT_OF_FLOAT_SCALE, "expect %s but got %s<%d,%d>", fs, v, w, p)
	}
	return nil
}

func (fs *FloatScaleVa) String() string {
	if fs.digital == nil && fs.decimel == nil {
		return ""
	}
	b := bytes.NewBuffer(nil)
	b.WriteByte('<')
	if fs.digital != nil {
		b.WriteString(fmt.Sprint(*fs.digital))
	}
	if fs.decimel != nil {
		b.WriteByte(',')
		b.WriteString(fmt.Sprint(*fs.decimel))
	}
	b.WriteByte('>')
	return b.String()
}
