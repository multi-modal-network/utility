package logic

import (
	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
	"onosutil/model"
	"onosutil/utils/calc"
	"onosutil/utils/errors"
	"onosutil/utils/format"
	"strconv"
	"strings"
)

// PrepareFlowsHandler 根据源目主机和模态类型计算需要下发流表的目标，返回（deviceID/port）结构数组，oar会知道怎么下发具体流表
func (m *Manager) PrepareFlowsHandler(ctx *context.Context) {
	srcHost, dstHost, modalType := ctx.Input.Query("src_host"), ctx.Input.Query("dst_host"), ctx.Input.Query("modal_type")
	if srcHost == "" || dstHost == "" || modalType == "" {
		log.Error("PrepareFlowsHandler invalid params")
		responseError(ctx, errors.PrepareFlowFailed)
		return
	}
	src, err := strconv.ParseInt(srcHost, 10, 64)
	if err != nil {
		log.Error("PrepareFlowsHandler src_host parse failed")
		responseError(ctx, err)
		return
	}
	dst, err := strconv.ParseInt(dstHost, 10, 64)
	if err != nil {
		log.Error("PrepareFlowsHandler dst_host parse failed")
		responseError(ctx, err)
		return
	}
	devices := calc.GetPathDevices(int32(src), int32(dst))
	log.Infof("PrepareFlowsHandler getPathInfo devices: %v", devices)
	flows, reachable := make([]string, 0), true
	for _, dev := range devices {
		if reachable == false {
			break
		}
		// check pipeconf
		device := model.Device{}
		if err := m.db.QueryTable(&model.Device{}).Filter("device_name__exact", dev.DeviceName).One(&device); err != nil {
			log.Warnf("PrepareFlowsHandler path device not found, err: %v", err)
			reachable = false
			continue
		}
		mode := format.ModelStringCorrect(modalType)
		if !strings.Contains(device.SupportModal, mode) {
			log.Warnf("PrepareFlowsHandler device %v pipeconf not support", dev.DeviceName)
			reachable = false
			continue
		}
		// 更新flows
		flows = append(flows, strings.Join(append([]string{}, strings.ToLower(device.DeviceID), strconv.Itoa(int(dev.Port))), "/"))
	}
	log.Infof("PrepareFlowsHandler flows: %v", flows)
	flowsStr := strings.Join(flows, ",")
	responseSuccess(ctx, flowsStr)
}
