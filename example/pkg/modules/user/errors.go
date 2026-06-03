package user

import "net/http"

type ErrorForbidden struct{}

func (ErrorForbidden) Error() string {
	return "权限不足"
}

func (ErrorForbidden) StatusCode() int {
	return http.StatusForbidden
}

type ErrorNotFound struct{}

func (ErrorNotFound) Error() string {
	return "用户不存在"
}

func (ErrorNotFound) StatusCode() int {
	return http.StatusNotFound
}
