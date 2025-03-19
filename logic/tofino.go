package logic

import (
	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
	"onosutil/model"
)

// TofinoPortResponse 模态转发端口 回包
type TofinoPortResponse struct {
	Port    int32 `json:"port"`
	Changed bool  `json:"changed"`
}

// GetTofinoPortHandler A->C（C->A）区通信，获取模态对应的TofinoA（TofinoC）转发端口
func (m *Manager) GetTofinoPortHandler(ctx *context.Context) {
	switchID := ctx.Input.Query("switchID")
	modalType := ctx.Input.Query("modalType")
	tofino := &model.TofinoPort{}
	if err := m.db.QueryTable(&model.TofinoPort{}).Filter("switch_id__exact", switchID).
		Filter("modal_type__exact", modalType).One(tofino); err != nil {
		log.Errorf("GetTofinoPortHandler query port failed: %v", err)
		responseError(ctx, err)
		return
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
