package va

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation/rule"
)

func NewIntScaleVa(r rule.Rule) (*IntScaleVa, error) {
	is := &IntScaleVa{
		unsigned: strings.HasPrefix(r.Name(), "uint"),
		bits:     32,
	}
	prefix := "int"
	if is.unsigned {
		prefix = "uint"
	}

	if suffix := strings.TrimPrefix(r.Name(), prefix); len(suffix) > 0 {
		bits, err := strconv.ParseUint(suffix, 10, 8)
		if err != nil || bits > 64 {
			return nil, codex.Errorf(ERROR__INVALID_INT_BITS, "expect bits in [1,64] got %s", r.Name())
		}
		is.bits = uint8(bits)
	} else {
		params := r.Parameters()
		if len(params) != 1 && len(params) != 0 {
			return nil, codex.Errorf(ERROR__INVALID_INT_BITS, "got %d parameters", len(params))
		}
		if len(params) == 1 {
			lit, ok := params[0].(*rule.Literal)
			if !ok {
				return nil, codex.Errorf(ERROR__INVALID_INT_BITS, "got parameter %s", params[0].Type())
			}
			bits, err := strconv.ParseUint(lit.String(), 10, 8)
			if err != nil || bits > 64 {
				return nil, codex.Errorf(ERROR__INVALID_INT_BITS, "got %s", r.Name())
			}
			is.bits = uint8(bits)
		}
	}
	is.mini, is.maxi = -1<<(is.bits-1), 1<<(is.bits-1)-1
	is.minu, is.maxu = 0, uint64(1<<is.bits)

	return is, nil
}

type IntScaleVa struct {
	unsigned   bool
	bits       uint8
	mini, maxi int64
	minu, maxu uint64
}

func (is *IntScaleVa) BuiltTo(b rule.Builder) {
	if is.unsigned {
		b.SetName("uint")
	} else {
		b.SetName("int")
	}
	b.AddParameters(rule.NewLiteral(strconv.Itoa(int(is.bits))))
}

func (is *IntScaleVa) Unsigned() bool {
	return is.unsigned
}

func (is *IntScaleVa) Bits() uint8 {
	return is.bits
}

func (is *IntScaleVa) String() string {
	if is.unsigned {
		return fmt.Sprintf("[%d,%d](unsigned %d bits)", is.mini, is.maxi, is.bits)
	}
	return fmt.Sprintf("[%d,%d](signed %d bits)", is.minu, is.maxu, is.bits)
}

func (is *IntScaleVa) Validate(v any) error {
	ok := false
	switch x := v.(type) {
	case int64:
		ok = !is.unsigned && is.mini <= x && x <= is.maxi
	case uint64:
		ok = is.unsigned && is.minu <= x && x <= is.maxu
	default:
		return codex.Errorf(ERROR__INVALID_INT_VALUE, "got %v(%T)", v, v)
	}
	if ok {
		return nil
	}
	return codex.Errorf(ERROR__OUT_OF_INT_BITS, "expect %s got %v(%T)", is, v, v)
}
