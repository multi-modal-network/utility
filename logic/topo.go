package logic

import (
	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
	"onosutil/model"
	"onosutil/utils/calc"
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

func (m *Manager) UpdateTopoHandler(ctx *context.Context) {
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
	// 包
	responseSuccess(ctx, TopoResponse{
		devNumber:  int(devNum),
		linkNumber: int(linkNum),
	})
}
