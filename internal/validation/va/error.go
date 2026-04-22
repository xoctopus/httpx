package va

// Error va error codes
// +genx:code
type Error int8

const (
	ERROR_UNDEFINED                 Error = iota
	ERROR__INVALID_LENGTH_RANGE           // invalid length range, expect 1 or 2 uint values
	ERROR__OUT_OF_LENGTH                  // input value out of length limitation
	ERROR__INVALID_VALUE_RANGE            // invalid range value
	ERROR__OUT_OF_VALUE_RANGE             // input value out of value range
	ERROR__INVALID_ENUM                   // invalid enumeration value
	ERROR__OUT_OF_ENUMERATED_VALUES       // input value out of enumerated values
	ERROR__INVALID_MULTIPLE               // invalid multiple
	ERROR__NOT_MATCH_MULTIPLE             // input value not match multiple
	ERROR__INVALID_FLOAT_SCALE            // invalid float scale, expect most 2 uint max 31
	ERROR__OUT_OF_FLOAT_SCALE             // input value out of float scale
	ERROR__INVALID_INT_BITS               // invalid int bits, expect most 1 uint max 64
	ERROR__OUT_OF_INT_BITS                // input value out of int bits
	ERROR__NOT_MATCH_REGEXP               // input value not match regexp
	ERROR__INVALID_INT_VALUE              // invalid input, expect int64/uint64
)
