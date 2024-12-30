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

type TopoResponse struct {
	devNumber  int
	linkNumber int
}

// 向onos推送netcfg
func sendNetcfgToONOS(ctx *context.Context) error {
	url := "http://127.0.0.1:8181/onos/v1/network/configuration"
	// 去除json中的links字段内容（ONOS的API不识别）
	b := ctx.Input.RequestBody
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		log.Error("sendNetcfgToONOS json.Unmarshal err:", err)
		return err
	}
	delete(data, "links")
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("sendNetcfgToONOS json.Marshal err:", err)
		return err
	}
	// 创建http请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("sendNetcfgToONOS http.NewRequest error:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("onos", "rocks")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("sendNetcfgToONOS http.DefaultClient.Do error:", err)
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("sendNetcfgToONOS io.ReadAll error:", err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Error("sendNetcfgToONOS failed:", resp.StatusCode)
		return errors.New(resp.StatusCode, "sendNetcfgToONOS failed:"+string(body))
	}
	log.Info("sendNetcfgToONOS response success")
	return nil
}

func (m *Manager) UpdateTopoHandler(ctx *context.Context) {
	if err := sendNetcfgToONOS(ctx); err != nil {
		log.Error("sendNetcfgToONOS failed, error:", err)
		responseError(ctx, err)
		return
	}
	netcfg := NetConf{}
	if err := ctx.BindJSON(&netcfg); err != nil {
		responseError(ctx, err)
		return
	}
	// 设备信息入库
	var devices []model.Device
	for deviceID, _ := range netcfg.Devices {
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
		devices = append(devices, model.Device{
			DeviceID: deviceID,
			Domain:   domain,
			Group:    group,
			SwitchID: switchID,
		})
	}
	devNum, err := m.db.InsertMulti(len(devices), devices)
	if err != nil {
		responseError(ctx, err)
		return
	}
	// 链路信息入库
	var links []model.Link
	for _, link := range netcfg.Links {
		links = append(links, model.Link{
			EndPoint1: link.EndPoint1,
			EndPoint2: link.EndPoint2,
		})
	}
	linkNum, err := m.db.InsertMulti(len(links), links)
	if err != nil {
		responseError(ctx, err)
		return
	}
	// 回包
	responseSuccess(ctx, TopoResponse{
		devNumber:  int(devNum),
		linkNumber: int(linkNum),
	})
}
