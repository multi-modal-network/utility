package logic

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"onosutil/model"
	"onosutil/utils/calc"
	"onosutil/utils/errors"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
)

type NetConf struct {
	Devices map[string]DeviceBasic `json:"devices"`
	Links   map[string]LinkBasic   `json:"links"`
}

type DeviceBasic struct {
	Basic DeviceInfo `json:"basic"`
}

type LinkBasic struct {
	Basic struct{} `json:"basic"`
}

type DeviceInfo struct {
	ManagementAddress string `json:"managementAddress"`
	Driver            string `json:"driver"`
	Pipeconf          string `json:"pipeconf"`
}

// 回包
type updateTopoResponse struct {
	DevNumber  int64 `json:"devNumber"`
	LinkNumber int64 `json:"linkNumber"`
}

type getTopoResponse struct {
	Devices map[string]DeviceBasic `json:"devices"`
	Ports   map[string]struct{}    `json:"ports"`
	Apps    map[string]struct{}    `json:"apps"`
	Hosts   map[string]struct{}    `json:"hosts"`
	Layouts map[string]struct{}    `json:"layouts"`
	Links   map[string]LinkBasic   `json:"links"`
	Region  map[string]struct{}    `json:"regions"`
}

// 向onos推送netcfg
func sendNetcfgToONOS(ctx *context.Context) (time.Duration, error) {
	startTime := time.Now()
	// 三台onos的url
	onos1Url := "http://127.0.0.1:8184/onos/v1/network/configuration"
	onos2Url := "http://127.0.0.1:8182/onos/v1/network/configuration"
	onos3Url := "http://127.0.0.1:8183/onos/v1/network/configuration"
	// 去除json中的links字段内容（ONOS的API不识别）
	b := ctx.Input.RequestBody
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		elapsedTime := time.Since(startTime)
		log.Error("sendNetcfgToONOS json.Unmarshal err:", err)
		return elapsedTime, err
	}
	delete(data, "links")
	// 解析devices
	deviceRaw, ok := data["devices"].(map[string]interface{})
	if !ok {
		elapsedTime := time.Since(startTime)
		log.Error("devices field not found or invalid")
		return elapsedTime, errors.New(400, "devices field not found or invalid")
	}
	// 按照domain分组
	domainGroups := map[string]map[string]interface{}{
		"domain5": {},
		"domain7": {},
		"default": {},
	}
	for deviceID, device := range deviceRaw {
		if strings.Contains(deviceID, "domain5") {
			domainGroups["domain5"][deviceID] = device
		} else if strings.Contains(deviceID, "domain7") {
			domainGroups["domain7"][deviceID] = device
		} else {
			domainGroups["default"][deviceID] = device
		}
	}
	for domain, devices := range domainGroups {
		if len(devices) == 0 {
			continue
		}
		domainData := map[string]interface{}{
			"devices": devices,
		}
		jsonData, err := json.Marshal(domainData)
		if err != nil {
			elapsedTime := time.Since(startTime)
			log.Error("sendNetcfgToONOS json.Marshal err:", err)
			return elapsedTime, err
		}
		var url string
		if domain == "domain5" {
			url = onos2Url
		} else if domain == "domain7" {
			url = onos3Url
		} else {
			url = onos1Url
		}
		// 创建http请求
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			elapsedTime := time.Since(startTime)
			log.Error("sendNetcfgToONOS http.NewRequest error:", err)
			return elapsedTime, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("onos", "Fiberhome@2020")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			elapsedTime := time.Since(startTime)
			log.Error("sendNetcfgToONOS http.DefaultClient.Do error:", err)
			return elapsedTime, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			elapsedTime := time.Since(startTime)
			log.Error("sendNetcfgToONOS io.ReadAll error:", err)
			return elapsedTime, err
		}
		if resp.StatusCode != 200 {
			elapsedTime := time.Since(startTime)
			log.Error("sendNetcfgToONOS failed:", resp.StatusCode)
			return elapsedTime, errors.New(resp.StatusCode, "sendNetcfgToONOS failed:"+string(body))
		}
	}
	elapsedTime := time.Since(startTime)
	log.Info("sendNetcfgToONOS response success")
	return elapsedTime, nil
}

// UpdateTopoHandler 上传netcfg.json文件更新拓扑
func (m *Manager) UpdateTopoHandler(ctx *context.Context) {
	// 转发Netcfg至ONOS
	elapsedTime, err := sendNetcfgToONOS(ctx)
	if err != nil {
		log.Error("sendNetcfgToONOS failed, error:", err)
		responseError(ctx, err)
		return
	}
	log.Info("sendNetcfgToONOS elapsedTime: ", elapsedTime)
	// 处理拓扑信息
	netcfg := NetConf{}
	if err := ctx.BindJSON(&netcfg); err != nil {
		responseError(ctx, err)
		return
	}
	devNumber, linkNumber := int64(0), int64(0)
	// 设备信息入库
	devices := make([]model.Device, 0)
	for deviceID, d := range netcfg.Devices {
		deviceName, err := calc.ExtractDeviceName(deviceID)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: invalid deviceID: %s, %s", deviceID, err)
			continue
		}
		domain, err := calc.ExtractDomain(deviceID)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: invalid deviceID: %s, %s", deviceID, err)
			continue
		}
		group, err := calc.ExtractGroup(deviceID)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: invalid deviceID: %s, %s", deviceID, err)
			continue
		}
		switchID, err := calc.ExtractSwitchID(deviceID)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: invalid deviceID: %s, %s", deviceID, err)
			continue
		}
		supportModal, err := calc.ExtractSupportModal(d.Basic.Pipeconf)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: invalid pipeconf: %s, %s", d.Basic.Pipeconf, err)
			continue
		}
		devices = append(devices, model.Device{
			DeviceID:          deviceID,
			DeviceName:        deviceName,
			Domain:            domain,
			Group:             group,
			SwitchID:          switchID,
			ManagementAddress: d.Basic.ManagementAddress,
			Driver:            d.Basic.Driver,
			Pipeconf:          d.Basic.Pipeconf,
			SupportModal:      strings.ToLower(supportModal),
		})
	}
	if len(devices) != 0 {
		devNumber, err = m.db.InsertMulti(100, devices)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: device InsertMulti error: %s", err)
			responseError(ctx, err)
			return
		}
	}
	// 链路信息入库
	links := make([]model.Link, 0)
	for linkStr, _ := range netcfg.Links {
		parts := strings.Split(linkStr, "-")
		links = append(links, model.Link{
			EndPoint1: parts[0],
			EndPoint2: parts[1],
		})
	}
	if len(links) != 0 {
		linkNumber, err = m.db.InsertMulti(100, links)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: link InsertMulti error: %s", err)
			responseError(ctx, err)
			return
		}
	}
	// 回包
	ctx.JSONResp(updateTopoResponse{
		DevNumber:  devNumber,
		LinkNumber: linkNumber,
	})
}

// GetTopoHandler 获取拓扑
func (m *Manager) GetTopoHandler(ctx *context.Context) {
	var devices []model.Device
	if _, err := m.db.QueryTable(&model.Device{}).All(&devices); err != nil {
		log.Error("GetTopoHandler query devices error:", err)
		responseError(ctx, err)
		return
	}
	var links []model.Link
	if _, err := m.db.QueryTable(&model.Link{}).All(&links); err != nil {
		log.Error("GetTopoHandler query links error:", err)
		responseError(ctx, err)
		return
	}
	res := getTopoResponse{
		Devices: make(map[string]DeviceBasic),
		Ports:   make(map[string]struct{}),
		Apps:    make(map[string]struct{}),
		Hosts:   make(map[string]struct{}),
		Layouts: make(map[string]struct{}),
		Links:   make(map[string]LinkBasic),
		Region:  make(map[string]struct{}),
	}
	for _, d := range devices {
		basic := DeviceBasic{Basic: DeviceInfo{
			Driver:            d.Driver,
			ManagementAddress: d.ManagementAddress,
			Pipeconf:          d.Pipeconf,
		}}
		res.Devices[d.DeviceID] = basic
	}
	for _, l := range links {
		basic := LinkBasic{Basic: struct{}{}}
		link := strings.Join(append([]string{}, l.EndPoint1, l.EndPoint2), "-")
		res.Links[link] = basic
	}
	ctx.JSONResp(res)
}
