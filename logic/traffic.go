package logic

import (
	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
	"onosutil/model"
	"onosutil/utils/calc"
	"onosutil/utils/format"
	"strconv"
	"strings"
	"time"
)

// TrafficInfo 流量三元组信息和发包时间
type TrafficInfo struct {
	DateTime int64  `json:"datetime"`
	SrcHost  int32  `json:"src_host"`
	DstHost  int32  `json:"dst_host"`
	ModeName string `json:"mode_name"`
}

type TrafficResponse struct {
	SrcHost  int32
	DstHost  int32
	ModeName string
	DateTime time.Time
	PathInfo []string
}

// RecordTrafficHandler 流量记录
func (m *Manager) RecordTrafficHandler(ctx *context.Context) {
	trafficInfo := &TrafficInfo{}
	if err := ctx.BindJSON(&trafficInfo); err != nil {
		responseError(ctx, err)
		return
	}
	// 获取理论上最佳路径
	devices := calc.GetPathDevices(trafficInfo.SrcHost, trafficInfo.DstHost)
	log.Infof("devices: %v", devices)
	// 获取实际pathInfo （流量可能被理论路径上的某个交换机截断，原因：转发端口不存在、pipeconf不支持模态）
	var pathInfo []string
	reachable := true
	for _, dev := range devices {
		if reachable == false {
			break
		}
		// check 转发端口
		if dev.Port == 0 {
			switchID := calc.GetSwitchID(dev.DeviceID)
			tofino := &model.TofinoPort{}
			if err := m.db.QueryTable(&model.TofinoPort{}).Filter("switch_id__exact", switchID).
				Filter("modal_type__exact", trafficInfo.ModeName).One(tofino); err != nil {
				log.Warnf("RecordTrafficHandler device %v not support", dev.DeviceID)
				reachable = false
				continue
			}
		}
		// check pipeconf
		device := model.Device{}
		log.Info("deviceid:%s", dev.DeviceID)
		if err := m.db.QueryTable(&model.Device{}).Filter("device_id", dev.DeviceID).One(&device); err != nil {
			log.Errorf("RecordTrafficHandler path device not found, err: %v", err)
			return
		}
		mode := format.ModelStringCorrect(trafficInfo.ModeName)
		if !strings.Contains(device.SupportModal, mode) {
			log.Warnf("RecordTrafficHandler device %v not support", dev.DeviceID)
			reachable = false
			continue
		}
		// 更新pathInfo
		pathInfo = append(pathInfo, strings.Join(append([]string{}, dev.DeviceID, strconv.Itoa(int(dev.Port))), "/"))
	}
	traffic := model.TrafficHistory{
		SrcHost:  trafficInfo.SrcHost,
		DstHost:  trafficInfo.DstHost,
		ModeName: trafficInfo.ModeName,
		Datetime: time.Unix(trafficInfo.DateTime, 0),
		PathInfo: strings.Join(pathInfo, ","),
	}
	if _, err := m.db.Insert(&traffic); err != nil {
		log.Errorf("RecordTrafficHandler err: %v", err)
		return
	}
	responseSuccess(ctx, nil)
}

// QueryTrafficHandler 流量查询
func (m *Manager) QueryTrafficHandler(ctx *context.Context) {
	// todo: 分页查询
	var traffics []model.TrafficHistory
	if _, err := m.db.QueryTable(&model.TrafficHistory{}).All(&traffics); err != nil {
		log.Error("QueryTrafficHandler query error:", err)
		responseError(ctx, err)
		return
	}
	res := make([]TrafficResponse, len(traffics))
	for _, t := range traffics {
		res = append(res, TrafficResponse{
			SrcHost:  t.SrcHost,
			DstHost:  t.DstHost,
			ModeName: t.ModeName,
			DateTime: t.Datetime,
			PathInfo: strings.Split(t.PathInfo, ","),
		})
	}
	responseSuccess(ctx, res)
}
