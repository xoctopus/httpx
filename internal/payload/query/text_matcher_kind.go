package query

type TextMatcherKind int

const (
	TEXT_MATCHER_KIND_UNKNOWN TextMatcherKind = 0
	TEXT_MATCHER_KIND__STRICT
	TEXT_MATCHER_KIND__CONTAIN
	TEXT_MATCHER_KIND__PREFIX
	TEXT_MATCHER_KIND__SUFFIX
)

type RangeMatcherKind int
