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
	Devices map[string]DeviceBasic `json:"devices"`
	Links   []Links                `json:"links"`
}

type Links struct {
	EndPoint1 string `json:"endpoint1"`
	EndPoint2 string `json:"endpoint2"`
}

type DeviceBasic struct {
	Basic DeviceInfo `json:"basic"`
}

type DeviceInfo struct {
	ManagementAddress string `json:"managementAddress"`
	Driver            string `json:"driver"`
	Pipeconf          string `json:"pipeconf"`
}

// 回包
type updateTopoResponse struct {
	DevNumber  int
	LinkNumber int
}

type getTopoResponse struct {
	Devices []model.Device `json:"devices"`
	Links   []model.Link   `json:"links"`
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
	// 设备信息入库
	var devices []model.Device
	if _, err := m.db.QueryTable(&model.Device{}).All(&devices); err != nil {
		log.Error("UpdateTopoHandler query devices error:", err)
		responseError(ctx, err)
		return
	}
	deviceIDMapping := make(map[string]interface{}, len(devices))
	for _, device := range devices {
		deviceIDMapping[device.DeviceID] = device
	}
	for deviceID, d := range netcfg.Devices {
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
		device := model.Device{
			DeviceID:          deviceID,
			Domain:            domain,
			Group:             group,
			SwitchID:          switchID,
			ManagementAddress: d.Basic.ManagementAddress,
			Driver:            d.Basic.Driver,
			Pipeconf:          d.Basic.Pipeconf,
			SupportModal:      supportModal,
		}
		// 不存在，insert；存在，update
		if _, ok := deviceIDMapping[deviceID]; !ok {
			_, err := m.db.Insert(&device)
			if err != nil {
				log.Error("UpdateTopoHandler insert device failed: ", device)
				responseError(ctx, err)
				return
			}
		} else {
			_, err := m.db.Update(&device)
			if err != nil {
				log.Error("UpdateTopoHandler update device failed: ", device)
				responseError(ctx, err)
				return
			}
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
		LinkMapping[linkStr] = link
	}
	for _, l := range netcfg.Links {
		linkStr := l.EndPoint1 + "/" + l.EndPoint2
		// 不存在，insert；存在，update
		link := model.Link{
			EndPoint1: l.EndPoint1,
			EndPoint2: l.EndPoint2,
		}
		if _, ok := LinkMapping[linkStr]; !ok {
			_, err := m.db.Insert(&link)
			if err != nil {
				log.Error("UpdateTopoHandler insert link failed: ", l.EndPoint1, l.EndPoint2)
				responseError(ctx, err)
			}
		} else {
			_, err := m.db.Update(&link)
			if err != nil {
				log.Error("UpdateTopoHandler update link failed: ", l.EndPoint1, l.EndPoint2)
				responseError(ctx, err)
			}
		}
	}

	responseSuccess(ctx, updateTopoResponse{
		DevNumber:  len(netcfg.Devices),
		LinkNumber: len(netcfg.Links),
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
	responseSuccess(ctx, getTopoResponse{
		Devices: devices,
		Links:   links,
	})
}
