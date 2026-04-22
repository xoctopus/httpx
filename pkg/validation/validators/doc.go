// Package validators defines field validation
//
// For basic types, such as integer, float, string, boolean, regexp
// For compose types with basic type elements, map, slice
//
// refs:
// JavaScript safe number
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MIN_SAFE_INTEGER
// IEEE 754
// https://en.wikipedia.org/wiki/IEEE_754
//
// Usage:
//
//	@RULE_NAME<RULE_PARAMETER_LIST>[RULE_VALUE_RANGE]{RULE_ENUMERATION_LIST}RULE_REQUIREMENTS
//
//	1. RULE_NAME: literal composed by ASCII, '-', '_'. MUST be prefixed with '@'
//	2. RULE_PARAMETERS: MUST is a rule or literal, use ',' to seperate parameters.
//		eg:
//			@float<22,4>       => float value with max 22 digitals and 4 decimal part.
//			@slice<@int>       => int slice
//			@map<@string,@int> => map with string key and int value
//	3. RULE_VALUE_RANGE: use '[',']','(',')'.
//		eg:
//			[,10]    => GET type minimum and LET 10
//			[10,15]  => GET 10 and LET 15
//			(,10)    => GT type minimum and LT 10
//			(10,15)  => GT 10 and LT 15
//			[10]     => elements MUST equal 10
//	4. RULE_ENUMERATION_LIST: use '{}' quoted and ',' spliter. basic type only, composed type(slice,map) is not supported.
//		eg:
//			@int{1,2,3}         => int value must in 1,2,3
//			@string{abc,def,gh} => string value must in abc,def,gh
//			@slice<@int>{1,2,3} => unsupported
//	5. RULE_REQUIREMENTS: suffix with '?' means value is optional, use '=' means value has default
//		eg:
//			@int        => int value and MUST not be 0
//			@string     => string value and MUST not be empty
//			@int?       => optional int value
//			@string?    => optional string value
//			@int=1      => optional int value with 1 as its default
//			@float=0.1  => optional float value with 0.1 as its default
//			@string=ABC => optional string value with 'ABC' as its default
//
// Int
//
//	@int                    => value MUST between minimum int32 and maximum int32
//	@int<8>                 => value MUST be with max 8 bits
//	@int<53>                => value MUST be with max 53 bits(js safe integer)
//	@int[1,10)              => value MUST be GET 1 and LT 10
//	@int[1,)                => value MUST be GET 1 and LT the maximum of int32
//	@int(,1)                => value MUST be LT 1 and GT minimum of int32
//	@int{1,2,3}             => value MUST be one of {1,2,3}
//	@int{%2}                => value MUST be multiple of 2. multiple mode MUST contain only 1 parameter.
//	@int<8>                 => be equivalent of @int8
//	@int<16>                => be equivalent of @int16
//	@int<32>                => be equivalent of @int32
//	@int<64>                => be equivalent of @int64
//	@uint<64>               => be equivalent of @uint64
//
// Float
//
//	@float[1,10)            => value MUST be larger than or equal to 1 and less than 10
//	@float<5,2>             => value width MUST be LET 5 and precision(scale) MUST be LET 2. eg: 11.123 -> too many fractional 1234.56 too many digits
//	@float32                => be equivalent of @float
//	@float64                => be equivalent of @double
//
// String
//
//	@string/\w+/            => value MUST match \w+. regexp start and end with '/'
//	@string[1,5)            => value byte length MUST GET 1 and LT 5
//	@string<byte>[1,5]      => be equivalent of @string[1,5]
//	@string<rune>[1,5]      => value rune length MUST GET 1 and LET 5
//	@string{ab,cd,ef}       => value MUST be one of "ab", "de" and "ef"
//	@string(100)            => failed to parse, length mode must be quoted with '[]'
//
// Slice
//
//	@slice<@float32>         => slice MUST be with float32 elements
//	@slice<@int8{1,2,3}>[5]  => slice MUST be with int8 elements and length MUST be 5, and each element MUST be one of {1,2,3}
//
// Map
//
//		@map<@string{A,B,C},@int>[,100]
//	                          => map MUST be with string key and key MUST be one
//	                             of A,B,C and MUST be with int value
//	                             and map length MUST between 0 and 100
package validators
