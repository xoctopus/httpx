package confhttp

// Error confhttp error
// +genx:code
type Error int8

const (
	ERROR_UNDEFINED Error = iota
	ERROR__HOST_ALIAS_INVALID_INPUT
	ERROR__HOST_ALIAS_INVALID_IPV6_ADDR
)
