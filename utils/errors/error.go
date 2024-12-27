package errors

import (
	"fmt"
	"strconv"
)

// Error 通用错误体结构
type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e Error) Error() string {
	return strconv.Itoa(e.Code) + ": " + e.Msg
}

// New 创建新的Error
func New(code int, msg string) Error {
	return Error{
		Code: code,
		Msg:  msg,
	}
}

// Newf 格式化方式创建新Error
func Newf(code int, format string, args ...interface{}) Error {
	return New(code, fmt.Sprintf(format, args...))
}
