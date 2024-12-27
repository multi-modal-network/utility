package logic

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
	"onosutil/utils/errors"
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
		return nil, errors.SetupDBFailed
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
	errT := errors.Unknown
	if err != nil {
		errT = errors.New(errors.CodeInner, err.Error())
	}
	log.Errorf("API(%s %s) error: %s", ctx.Request.Method, ctx.Request.URL.Path, err.Error())
	ctx.Output.SetStatus(500)
	ctx.JSONResp(errors.Error{
		Code: errT.Code,
		Msg:  errT.Msg,
	})
}
