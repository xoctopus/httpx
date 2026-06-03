package va

import (
	"fmt"
	"regexp"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation/rule"
)

func NewRegexpVa(r rule.Rule, hint string) *RegexpVa {
	if len(r.Pattern()) == 0 {
		return nil
	}
	return &RegexpVa{
		pattern: r.Pattern(),
		regexp:  r.Regexp(),
		hint:    hint,
	}
}

type RegexpVa struct {
	regexp  *regexp.Regexp
	pattern string
	hint    string
}

func (v *RegexpVa) Pattern() string {
	return v.pattern
}

func (v *RegexpVa) Hint() string {
	return v.hint
}

func (v *RegexpVa) Validate(s string) error {
	if v != nil && !v.regexp.MatchString(s) {
		hint := ""
		if len(v.hint) > 0 {
			hint = fmt.Sprintf(" [hint: %s]", v.hint)
		}
		return codex.Errorf(ERROR__NOT_MATCH_REGEXP, "expect match /%s/ got %s%s", v.pattern, s, hint)
	}
	return nil
}

func (v *RegexpVa) BuiltTo(b rule.Builder) {
	if v != nil {
		_ = b.SetPattern(v.pattern)
	}
}
