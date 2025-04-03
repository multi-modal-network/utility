package logic

import (
	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
	"onosutil/model"
	"onosutil/utils/errors"
)

// GetTofinoPortHandler A->C（C->A）区通信，获取模态对应的TofinoA（TofinoC）转发端口
func (m *Manager) GetTofinoPortHandler(ctx *context.Context) {
	switchID := ctx.Input.Query("switchID")
	modalType := ctx.Input.Query("modalType")
	if switchID == "" || modalType == "" {
		log.Errorf("GetTofinoPortHandler empty query")
		responseError(ctx, errors.New(errors.CodeGetTofinoPortFailed, "GetTofinoPortHandler empty query"))
		return
	}
	tofino := &model.TofinoPort{}
	if err := m.db.QueryTable(&model.TofinoPort{}).Filter("switch_id__exact", switchID).
		Filter("modal_type__exact", modalType).One(tofino); err != nil {
		log.Errorf("GetTofinoPortHandler query port failed: %v", err)
		responseError(ctx, err)
		return
	}
	// TofinoPortResponse 模态转发端口 回包
	type TofinoPortResponse struct {
		Port    int32 `json:"port"`
		Changed bool  `json:"changed"`
	}
	res := TofinoPortResponse{
		Port:    tofino.Port,
		Changed: (tofino.OldPort == 0) || (tofino.OldPort != tofino.Port),
	}
	// 更新old_port
	tofino.OldPort = tofino.Port
	if _, err := m.db.Update(tofino, "old_port"); err != nil {
		log.Errorf("UpdateTopoHandler update old_port failed: %v", err)
		responseError(ctx, err)
		return
	}
	responseSuccess(ctx, res)
}

// ModifyTofinoPortHandler 修改Tofino交换机模态对应的转发端口
func (m *Manager) ModifyTofinoPortHandler(ctx *context.Context) {
	var req struct {
		SwitchID  int32  `json:"switchID"`
		ModalType string `json:"modalType"`
		Port      int32  `json:"port"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		responseError(ctx, err)
		return
	}
	tofino := &model.TofinoPort{}
	if err := m.db.QueryTable(&model.TofinoPort{}).Filter("switch_id__exact", req.SwitchID).
		Filter("modal_type__exact", req.ModalType).One(tofino); err != nil {
		log.Errorf("ModifyTofinoPortHandler query port failed: %v", err)
		responseError(ctx, err)
		return
	}
	tofino.OldPort = tofino.Port
	tofino.Port = req.Port
	if _, err := m.db.Update(tofino, "old_port"); err != nil {
		log.Errorf("ModifyTofinoPortHandler update failed: %v", err)
		responseError(ctx, err)
		return
	}
	responseSuccess(ctx, nil)
}
