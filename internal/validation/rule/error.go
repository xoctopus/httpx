package rule

// Error defines error code for validation rule
// +genx:code
type Error int8

const (
	ERROR_UNDEFINED                      Error = iota
	ERROR__INVALID_RULE_LEADER                 // invalid rule name leader, expect start with '@'
	ERROR__MISSING_RULE_NAME                   // missing rule name
	ERROR__INVALID_PARAM_LEADER                // invalid rule parameter list leader, expect start with '<'
	ERROR__INVALID_PARAM_TAILER                // invalid rule parameter list tailer, expect end with '>'
	ERROR__INVALID_PARAMS                      // invalid rule parameter list content
	ERROR__INVALID_RANGE_LEADER                // invalid rule range leader, expect start with '[' or '('
	ERROR__INVALID_RANGE_TAILER                // invalid rule range tailer, expect end with ']' or ')'
	ERROR__INVALID_RANGE_COUNT                 // invalid rule range count, expect 1 or 2
	ERROR__INVALID_RANGE_MODE                  // invalid rule range mode, while length mode must be quoted with '[]'
	ERROR__INVALID_VALUE_LEADER                // invalid rule value leader, expect start with '{'
	ERROR__INVALID_VALUE_TAILER                // invalid rule value tailer, expect end with '}'
	ERROR__INVALID_REGEXP_LEADER               // invalid rule matcher leader, expect start with '/'
	ERROR__INVALID_REGEXP_TAILER               // invalid rule matcher leader, expect end with '/'
	ERROR__FAILED_COMPILE_REGEXP               // failed to compile regexp
	ERROR__INVALID_QUOTED_DEFAULT_TAILER       // invalid rule quoted default value tailer, expect quoted with `'`
	ERROR__INVALID_LITERAL                     // invalid rule literal
)
