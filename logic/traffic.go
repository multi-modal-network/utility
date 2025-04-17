package logic

import (
	"fmt"
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
	SrcHost  string    `json:"srcHost"`
	DstHost  string    `json:"dstHost"`
	ModeName string    `json:"modeName"`
	DateTime time.Time `json:"datetime"`
	PathInfo []string  `json:"pathInfo"`
}

// RecordTrafficHandler 流量记录
func (m *Manager) RecordTrafficHandler(ctx *context.Context) {
	trafficInfo := TrafficInfo{}
	if err := ctx.BindJSON(&trafficInfo); err != nil {
		responseError(ctx, err)
		return
	}
	// 理论计算端到端通信的路径上所包含的交换机
	devices := calc.GetPathDevices(trafficInfo.SrcHost, trafficInfo.DstHost)
	log.Infof("RecordTrafficHandler getPathInfo devices: %v", devices)
	// 获取实际pathInfo （流量可能被路径上的某个交换机截断，原因：转发端口不存在、pipeconf不支持模态）
	pathInfo, reachable := make([]string, 0), true
	for _, dev := range devices {
		if reachable == false {
			break
		}
		// check pipeconf
		device := model.Device{}
		if err := m.db.QueryTable(&model.Device{}).Filter("device_name__exact", dev.DeviceName).One(&device); err != nil {
			log.Warnf("RecordTrafficHandler path device %s not found, err: %v", dev.DeviceName, err)
			reachable = false
			continue
		}
		mode := format.ModelStringCorrect(trafficInfo.ModeName)
		if !strings.Contains(device.SupportModal, mode) {
			log.Warnf("RecordTrafficHandler device %v pipeconf not support", dev.DeviceName)
			reachable = false
			continue
		}
		// 更新pathInfo
		pathInfo = append(pathInfo, strings.Join(append([]string{}, dev.DeviceName, strconv.Itoa(int(dev.Port))), "/"))
	}
	log.Infof("Practical Routing Path:%v", pathInfo)
	traffic := model.TrafficHistory{
		SrcHost:   trafficInfo.SrcHost,
		DstHost:   trafficInfo.DstHost,
		ModeName:  trafficInfo.ModeName,
		Timestamp: trafficInfo.DateTime,
		Datetime:  time.Unix(trafficInfo.DateTime, 0),
		PathInfo:  strings.Join(pathInfo, ","),
	}
	if _, err := m.db.Insert(&traffic); err != nil {
		log.Errorf("RecordTrafficHandler err: %v", err)
		return
	}
	responseSuccess(ctx, nil)
}

// QueryTrafficHandler 流量查询
func (m *Manager) QueryTrafficHandler(ctx *context.Context) {
	var traffics []model.TrafficHistory
	qs := m.db.QueryTable(&model.TrafficHistory{})
	t := ctx.Input.Query("time")
	if t != "" {
		timestamp, err := strconv.ParseInt(t, 0, 64)
		if err != nil {
			log.Errorf("QueryTrafficHandler parameter error, err: %v", err)
			responseError(ctx, err)
			return
		}
		qs = qs.Filter("timestamp__gte", timestamp)
	}
	if _, err := qs.All(&traffics); err != nil {
		log.Error("QueryTrafficHandler query error:", err)
		responseError(ctx, err)
		return
	}
	res := make([]TrafficResponse, 0)
	for _, traffic := range traffics {
		res = append(res, TrafficResponse{
			SrcHost:  fmt.Sprintf("h%d", traffic.SrcHost),
			DstHost:  fmt.Sprintf("h%d", traffic.DstHost),
			ModeName: traffic.ModeName,
			DateTime: traffic.Datetime,
			PathInfo: strings.Split(traffic.PathInfo, ","),
		})
	}
	ctx.JSONResp(res)
}
