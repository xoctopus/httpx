package validation

// Error validation error
// +genx:code
type Error int8

const (
	ERROR_UNDEFINED           Error = iota
	ERROR__STRING_LENGTH_MODE       // invalid string length mode, expect 'byte' or 'rune'
	ERROR__NOT_MATCH_REGEXP         // invalid string not match regexp
	ERROR__SLICE_PARAM              // invalid string parameter
	ERROR__MAP_PARAM                // invalid map parameter
	ERROR__INPUT_TYPE               // invalid input: invalid type
	ERROR__INPUT_VALUE              // invalid input: invalid value
	ERROR__MISSING_REQUIRED         // invalid input: missing required
	ERROR__UNREGISTERED_RULE        // unregistered rule
	ERROR__DEC_INVALID_INPUT        // decoder: invalid input, expecting a pointer value
)
