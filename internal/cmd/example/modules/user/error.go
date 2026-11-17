package user

// Error 用户模块错误码
// +genx:code
// @code domain=USER
type Error int

const (
	ERROR_UNDEFINED Error = iota
	// ERROR__NOT_FOUND_BY_NAME
	// @code api=StatusNotFound
	// @code hint=用户名不存在
	ERROR__NOT_FOUND_BY_NAME
	// ERROR__NOT_FOUND_BY_ID
	// @code api=StatusNotFound
	// @code hint=用ID不存在
	ERROR__NOT_FOUND_BY_ID
	// ERROR__REGISTRATION_NOT_OPEN
	// @code api=StatusForbidden
	// @code hint=注册未开放
	ERROR__REGISTRATION_NOT_OPEN
	// ERROR__USERNAME_CONFLICT
	// @code api=StatusConflict
	// @code hint=用户名已注册
	ERROR__USERNAME_CONFLICT
	// ERROR__PASSWORD_UNMATCHED
	// @code api=StatusUnauthorized
	// @code hint=登陆密码错误
	ERROR__PASSWORD_UNMATCHED
	// ERROR__FAILED_TO_REFRESH_TOKEN
	// @code api=StatusInternalServerError
	ERROR__FAILED_TO_REFRESH_TOKEN
)
