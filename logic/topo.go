package logic

import (
	"bytes"
	"encoding/json"
	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"onosutil/model"
	"onosutil/utils/calc"
	"onosutil/utils/errors"
	"time"
)

type NetConf struct {
	Devices map[string]Devices `json:"devices"`
	Links   []Links            `json:"links"`
}

type Links struct {
	EndPoint1 string `json:"endpoint1"`
	EndPoint2 string `json:"endpoint2"`
}

type Devices struct {
	Device map[string]DeviceBasic
}

type DeviceBasic struct {
	Basic map[string]DeviceInfo `json:"basic"`
}

type DeviceInfo struct {
	ManagementAddress string `json:"managementAddress"`
	Driver            string `json:"driver"`
	Pipeconf          string `json:"pipeconf"`
}

// 向onos推送netcfg
func sendNetcfgToONOS(ctx *context.Context) (time.Duration, error) {
	startTime := time.Now()
	url := "http://127.0.0.1:8181/onos/v1/network/configuration"
	// 去除json中的links字段内容（ONOS的API不识别）
	b := ctx.Input.RequestBody
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		elapsedTime := time.Since(startTime)
		log.Error("sendNetcfgToONOS json.Unmarshal err:", err)
		return elapsedTime, err
	}
	delete(data, "links")
	jsonData, err := json.Marshal(data)
	if err != nil {
		elapsedTime := time.Since(startTime)
		log.Error("sendNetcfgToONOS json.Marshal err:", err)
		return elapsedTime, err
	}
	// 创建http请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		elapsedTime := time.Since(startTime)
		log.Error("sendNetcfgToONOS http.NewRequest error:", err)
		return elapsedTime, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("onos", "rocks")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		elapsedTime := time.Since(startTime)
		log.Error("sendNetcfgToONOS http.DefaultClient.Do error:", err)
		return elapsedTime, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
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
	elapsedTime := time.Since(startTime)
	log.Info("sendNetcfgToONOS response success")
	return elapsedTime, nil
}

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
	// 设备信息入库
	var devices []model.Device
	if _, err := m.db.QueryTable(&model.Device{}).All(&devices, "deviceID"); err != nil {
		log.Error("UpdateTopoHandler query devices error:", err)
		responseError(ctx, err)
		return
	}
	deviceIDMapping := make(map[string]interface{}, len(devices))
	for _, device := range devices {
		deviceIDMapping[device.DeviceID] = struct{}{}
	}
	updateDevices := make([]model.Device, 0, len(netcfg.Devices))
	for deviceID, _ := range netcfg.Devices {
		// 如果deviceID已存在，continue
		if _, ok := deviceIDMapping[deviceID]; ok {
			continue
		}
		domain, err := calc.ExtractDomain(deviceID)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: invalid deviceID: %s", deviceID)
			continue
		}
		group, err := calc.ExtractGroup(deviceID)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: invalid deviceID: %s", deviceID)
			continue
		}
		switchID, err := calc.ExtractSwitchID(deviceID)
		if err != nil {
			log.Errorf("UpdateTopoHandler error: invalid deviceID: %s", deviceID)
			continue
		}
		updateDevices = append(updateDevices, model.Device{
			DeviceID: deviceID,
			Domain:   domain,
			Group:    group,
			SwitchID: switchID,
		})
	}
	devNum := len(updateDevices)
	if devNum != 0 {
		_, err := m.db.InsertMulti(3, updateDevices)
		if err != nil {
			responseError(ctx, err)
			return
		}
	}
	// 链路信息入库
	links := make([]model.Link, 0, len(netcfg.Links))
	if _, err := m.db.QueryTable(&model.Link{}).All(&links); err != nil {
		log.Error("UpdateTopoHandler query links error:", err)
		responseError(ctx, err)
		return
	}
	LinkMapping := make(map[string]interface{}, len(links))
	for _, link := range links {
		linkStr := link.EndPoint1 + "/" + link.EndPoint2
		LinkMapping[linkStr] = struct{}{}
	}
	var updateLinks []model.Link
	for _, link := range netcfg.Links {
		linkStr := link.EndPoint1 + "/" + link.EndPoint2
		// 如果link已存在，continue
		if _, ok := LinkMapping[linkStr]; ok {
			continue
		}
		updateLinks = append(updateLinks, model.Link{
			EndPoint1: link.EndPoint1,
			EndPoint2: link.EndPoint2,
		})
	}
	linkNum := len(updateLinks)
	if linkNum != 0 {
		_, err := m.db.InsertMulti(3, updateLinks)
		if err != nil {
			responseError(ctx, err)
			return
		}
	}
	if err != nil {
		responseError(ctx, err)
		return
	}
	// 回包
	type updateResponse struct {
		DevNumber  int
		LinkNumber int
	}
	responseSuccess(ctx, updateResponse{
		DevNumber:  devNum,
		LinkNumber: linkNum,
	})
}
