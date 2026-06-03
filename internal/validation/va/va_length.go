package va

import (
	"bytes"
	"fmt"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation/rule"
)

func NewLengthVa(r rule.Rule) (*LengthVa, error) {
	if r.LengthMode() {
		if x, _ := r.Min(); !rule.IsNil(x) {
			length, err := ParseEnumValue[uint](x.String())
			if err != nil {
				return nil, codex.Wrapf(ERROR__INVALID_LENGTH_RANGE, err, "got %s", x)
			}
			return &LengthVa{length: new(length)}, nil
		}
		return nil, nil
	}

	minimum, exl := r.Min()
	maximum, exr := r.Max()
	if rule.IsNil(minimum) && rule.IsNil(maximum) {
		return nil, nil
	}

	lr := &LengthVa{min: 0, exMin: exl, exMax: exr}
	if !rule.IsNil(minimum) {
		v, err := ParseEnumValue[uint](minimum.String())
		if err != nil {
			return nil, codex.Wrapf(ERROR__INVALID_LENGTH_RANGE, err, "got %s", minimum)
		}
		lr.min = v
	}
	if !rule.IsNil(maximum) {
		v, err := ParseEnumValue[uint](maximum.String())
		if err != nil {
			return nil, codex.Wrapf(ERROR__INVALID_LENGTH_RANGE, err, "got %s", minimum)
		}
		lr.max = new(v)
		if lr.min >= *lr.max || lr.exMax && lr.min > *lr.max {
			return nil, codex.Errorf(ERROR__INVALID_LENGTH_RANGE, "min>max %d>%d", lr.min, *lr.max)
		}
	}
	return lr, nil
}

type LengthVa struct {
	min    uint
	exMin  bool
	max    *uint
	exMax  bool
	length *uint
}

func (lr *LengthVa) LengthRange() (*uint64, *uint64) {
	if lr == nil {
		return nil, nil
	}
	_max := (*uint64)(nil)
	if lr.max != nil {
		_max = new(uint64(*lr.max))
	}
	return new(uint64(lr.min)), _max
}

func (lr *LengthVa) String() string {
	if lr == nil {
		return ""
	}

	if lr.length != nil {
		return fmt.Sprintf("[%d]", *lr.length)
	}

	b := bytes.NewBuffer(nil)
	if lr.exMin {
		b.WriteByte('(')
	} else {
		b.WriteByte('[')
	}

	b.WriteString(fmt.Sprint(lr.min))
	b.WriteByte(',')
	if lr.max != nil {
		b.WriteString(fmt.Sprint(*lr.max))
	}

	if lr.exMax {
		b.WriteByte(')')
	} else {
		b.WriteByte(']')
	}

	return b.String()
}

func (lr *LengthVa) Validate(length uint) error {
	if lr == nil {
		return nil
	}
	if lr.length != nil {
		if length != *lr.length {
			return codex.Errorf(ERROR__OUT_OF_LENGTH, "expect equal %d got %d", *lr.length, length)
		}
		return nil
	}
	if (length < lr.min || lr.exMin && length <= lr.min) ||
		(lr.max != nil && (length > *lr.max || lr.exMax && length >= *lr.max)) {
		return codex.Errorf(ERROR__OUT_OF_LENGTH, "expect %s got %d", lr, length)
	}
	return nil
}

func (lr *LengthVa) BuiltTo(b rule.Builder) {
	if lr == nil {
		return
	}

	if lr.length != nil {
		b.SetLengthMode(true)
		b.SetMin(rule.NewLiteral(fmt.Sprint(*lr.length)), false)
		b.SetMax(rule.NewLiteral(""), false)
		return
	}

	_max := rule.NewLiteral("")
	if lr.max != nil {
		_max = rule.NewLiteral(fmt.Sprint(*lr.max))
	}
	b.SetMin(rule.NewLiteral(fmt.Sprint(lr.min)), lr.exMin)
	b.SetMax(_max, lr.exMax)

}
