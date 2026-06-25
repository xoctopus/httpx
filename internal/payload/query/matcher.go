package query

type TextMatcher interface {
	TextMatcherKey() string
}

type RangeMatcher interface {
	RangeMatcherKey() string
}
