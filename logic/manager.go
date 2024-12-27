package logic

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Manager 中心管理器，提供所有API
type Manager struct {
	db orm.Ormer
}

type Options struct {
	Ormer orm.Ormer
}

// 单例模式
var globalManager *Manager

func NewManager(option Options) (*Manager, error) {
	m := getManager()
	if m != nil {
		return m, nil
	}
	m = &Manager{
		db: option.Ormer,
	}
	// 参数检查
	if m.db == nil {
		return nil, errors.New("invalid Error: ormer is nil")
	}
	globalManager = m
	return m, nil
}

func getManager() *Manager {
	return globalManager
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 回包：成功
func responseSuccess(ctx *context.Context, data interface{}) {
	if data == nil {
		data = Response{
			Code: 0,
			Msg:  "success",
		}
	}
	ctx.JSONResp(Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

// 回包：错误
func responseError(ctx *context.Context, err error) {
	ctx.JSONResp(Response{
		Code: -1,
		Msg:  "error",
	})
}
