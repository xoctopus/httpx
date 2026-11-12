package status_test

import "fmt"

// UserCode is a domain code: every value declares a status code except
// USER__QUERY_FAILED, and USER__NO_HINT has no hint.
type UserCode int8

const (
	USER_UNDEFINED UserCode = iota
	USER__NOT_FOUND
	USER__QUERY_FAILED
	USER__NO_HINT
)

func (c UserCode) Message() string { return fmt.Sprintf("[USER:%d] message", c) }

func (c UserCode) StatusCode() int {
	switch c {
	case USER__NOT_FOUND:
		return 404
	case USER__NO_HINT:
		return 401
	default:
		return 0
	}
}

func (c UserCode) CodeIdentifier() string {
	switch c {
	case USER__NOT_FOUND:
		return "USER::NOT_FOUND"
	case USER__QUERY_FAILED:
		return "USER::QUERY_FAILED"
	case USER__NO_HINT:
		return "USER::NO_HINT"
	default:
		return "USER::UNDEFINED"
	}
}

func (c UserCode) CodeHint() string {
	switch c {
	case USER__NOT_FOUND:
		return "用户不存在"
	case USER__QUERY_FAILED:
		return "查询失败"
	default:
		return ""
	}
}

// SqlCode is a component code: it declares status codes but has no identifier
// and no hint.
type SqlCode int8

const (
	SQL__NOTFOUND SqlCode = iota + 1
	SQL__ROLLBACK
)

func (c SqlCode) Message() string { return fmt.Sprintf("[SQLERROR:%d] message", c) }

func (c SqlCode) StatusCode() int {
	switch c {
	case SQL__NOTFOUND:
		return 404
	default:
		return 500
	}
}
