package errors

// Error 通用错误体结构
type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
