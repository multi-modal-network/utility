package logic

import (
	"encoding/base64"
	"fmt"
	"onosutil/model"
	"onosutil/utils/calc"
	"onosutil/utils/errors"
	"onosutil/utils/format"
	"onosutil/utils/packetparse"
	postflow "onosutil/utils/postFlow"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
)


func parseModalParams[T packetparse.ModalParser](modalType string, bufferData []byte) (T, error) {
	var params T

	// 使用反射或类型断言来正确初始化指针类型
	switch any(params).(type) {
	case *packetparse.IDParams:
		idParams := &packetparse.IDParams{}
		res, err := idParams.Parse(bufferData)
		log.Infof("Parse %s params: %s", modalType, res)
		return any(idParams).(T), err
	case *packetparse.IPParams:
		ipParams := &packetparse.IPParams{}
		res, err := ipParams.Parse(bufferData)
		log.Infof("Parse %s params: %s", modalType, res)
		return any(ipParams).(T), err
	case *packetparse.NDNParams:
		ndnParams := &packetparse.NDNParams{}
		res, err := ndnParams.Parse(bufferData)
		log.Infof("Parse %s params: %s", modalType, res)
		return any(ndnParams).(T), err
	case *packetparse.GEOParams:
		geoParams := &packetparse.GEOParams{}
		res, err := geoParams.Parse(bufferData)
		log.Infof("Parse %s params: %s", modalType, res)
		return any(geoParams).(T), err
	case *packetparse.MFParams:
		mfParams := &packetparse.MFParams{}
		res, err := mfParams.Parse(bufferData)
		log.Infof("Parse %s params: %s", modalType, res)
		return any(mfParams).(T), err
	case *packetparse.FLEXIPParams:
		flexipParams := &packetparse.FLEXIPParams{}
		res, err := flexipParams.Parse(bufferData)
		log.Infof("Parse %s params: %s", modalType, res)
		return any(flexipParams).(T), err
	default:
		var zero T
		return zero, fmt.Errorf("unsupported type")
	}
}


type FlowRequest struct {
	SrcHost    int    `json:"src_host"`
	DstHost    int    `json:"dst_host"`
	ModalType  string `json:"modal_type"`
	BufferData string `json:"buffer_data"`
}

// PrepareFlowsHandler 根据源目主机和模态类型计算需要下发流表的目标，返回（deviceID/port）结构数组，oar会知道怎么下发具体流表
func (m *Manager) PrepareFlowsHandler(ctx *context.Context) {
	var req FlowRequest
	if err := ctx.BindJSON(&req); err != nil {
		log.Error("PrepareFlowsHandler read json failed: ", err)
		responseError(ctx, nil)
		return
	}
	if req.SrcHost == 0 || req.DstHost == 0 || req.ModalType == "" || req.BufferData == "" {
		log.Error("PrepareFlowsHandler invalid params")
		responseError(ctx, errors.PrepareFlowFailed)
		return
	}

	// base64解码bufferData
	bufferData, err := base64.StdEncoding.DecodeString(req.BufferData)
	if err != nil {
		log.Errorf("PrepareFlowsHandler base64 decode failed: %v", err)
		return
	}

	log.Infof("BufferData (Hex): % X", bufferData)

	var params packetparse.ModalParser
	// 解析对应模态下流表需要的参数
	switch req.ModalType {
	case "ipv4":
		params, err = parseModalParams[*packetparse.IPParams](req.ModalType, bufferData)
		if err != nil {
			log.Errorf("Parse IP params failed: %v", err)
			responseError(ctx, errors.PrepareFlowFailed)
			return
		}
		log.Infof("IP params: %+v", params)

	case "ndn":
		params, err = parseModalParams[*packetparse.NDNParams](req.ModalType, bufferData)
		if err != nil {
			log.Errorf("Parse NDN params failed: %v", err)
			responseError(ctx, errors.PrepareFlowFailed)
			return
		}
		log.Infof("NDN params: %+v", params)

	case "geo":
		params, err = parseModalParams[*packetparse.GEOParams](req.ModalType, bufferData)
		if err != nil {
			log.Errorf("Parse Geo params failed: %v", err)
			responseError(ctx, errors.PrepareFlowFailed)
			return
		}
		log.Infof("Geo params: %+v", params)
	case "mf":
		params, err = parseModalParams[*packetparse.MFParams](req.ModalType, bufferData)
		if err != nil {
			log.Errorf("Parse mf params failed: %v", err)
			responseError(ctx, errors.PrepareFlowFailed)
			return
		}
		log.Infof("mf params: %+v", params)
	case "flexip":
		params, err = parseModalParams[*packetparse.FLEXIPParams](req.ModalType, bufferData)
		if err != nil {
			log.Errorf("Parse flexip params failed: %v", err)
			responseError(ctx, errors.PrepareFlowFailed)
			return
		}
		log.Infof("flexip params: %+v", params)
	case "id":
		params, err = parseModalParams[*packetparse.IDParams](req.ModalType, bufferData)
		if err != nil {
			log.Errorf("Parse id params failed: %v", err)
			responseError(ctx, errors.PrepareFlowFailed)
			return
		}
		log.Infof("id params: %+v", params)

	default:
		log.Errorf("Unsupported modal type: %s", req.ModalType)
		responseError(ctx, errors.PrepareFlowFailed)
		return
	}
	log.Infof("PrepareFlowsHandler params: %+v", params)

	devices := calc.GetPathDevices(int32(req.SrcHost), int32(req.DstHost))
	log.Infof("PrepareFlowsHandler getPathInfo devices: %v", devices)
	flows, reachable := make([]string, 0), true

	for _, dev := range devices {
		if reachable == false {
			break
		}
		port := dev.Port
		// check 转发端口（Tofino交换机转发端口可能未确定）
		if dev.Port == 0 {
			switchID := calc.GetSwitchID(dev.DeviceName)
			tofino := &model.TofinoPort{}
			if err := m.db.QueryTable(&model.TofinoPort{}).Filter("switch_id__exact", switchID).
				Filter("modal_type__exact", req.ModalType).One(tofino); err != nil {
				log.Warnf("PrepareFlowsHandler device %v port not support", dev.DeviceName)
				reachable = false
				continue
			}
			port = tofino.Port
		}
		// check pipeconf
		device := model.Device{}
		if err := m.db.QueryTable(&model.Device{}).Filter("device_name__exact", dev.DeviceName).One(&device); err != nil {
			log.Warnf("PrepareFlowsHandler path device not found, err: %v", err)
			reachable = false
			continue
		}
		mode := format.ModelStringCorrect(req.ModalType)
		if !strings.Contains(device.SupportModal, mode) {
			log.Warnf("PrepareFlowsHandler device %v pipeconf not support", dev.DeviceName)
			reachable = false
			continue
		}
		// 更新flows
		flows = append(flows, strings.Join(append([]string{}, strings.ToLower(device.DeviceID), strconv.Itoa(int(port))), "/"))
	}

	// 分类
	domain5Flows, domain7Flows, defaultFlows := make([]string, 0), make([]string, 0), make([]string, 0)
	for _, flow := range flows {
		if strings.Contains(flow, "domain5") {
			domain5Flows = append(domain5Flows, flow)
		} else if strings.Contains(flow, "domain7") {
			domain7Flows = append(domain7Flows, flow)
		} else {
			defaultFlows = append(defaultFlows, flow)
		}
	}

	// 三台ONOS的流表下发URL
	onos1Url := "http://127.0.0.1:8181/onos/v1/flows?appId=org.stratumproject.basic-tna"
	onos2Url := "http://127.0.0.1:8182/onos/v1/flows?appId=org.stratumproject.basic-tna"
	onos3Url := "http://127.0.0.1:8183/onos/v1/flows?appId=org.stratumproject.basic-tna"

	_, err = postFlows(domain5Flows, onos2Url, req.ModalType, params)
	if err != nil {
		log.Errorf("PrepareFlowsHandler post domain5 flows failed: %v", err)
	}
	_, err = postFlows(domain7Flows, onos3Url, req.ModalType, params)
	if err != nil {
		log.Errorf("PrepareFlowsHandler post domain7 flows failed: %v", err)
	}
	_, err = postFlows(defaultFlows, onos1Url, req.ModalType, params)
	if err != nil {
		log.Errorf("PrepareFlowsHandler post default flows failed: %v", err)
	}
	responseSuccess(ctx, "")
}

func (m *Manager) AddNdnFlowHandler(ctx *context.Context) {
	cmd := exec.Command("sudo", "/bin/python3", "/ndn_flow.py")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Add ndn flow exec failed: %v\n output: %s", err, string(output))
		responseError(ctx, err)
		return
	}
	log.Infof("add ndn flow output:%s", string(output))
	responseSuccess(ctx, nil)
}

func postFlows(flows []string, url string, modalType string, params packetparse.ModalParser) (string, error) {
	switch modalType {
	case "flexip":
		flxipParam, ok := params.(*packetparse.FLEXIPParams)
		if !ok {
			return "", errors.InvalidParam
		}
		return postflow.ApplyFlexIPFlow(flows, url, *flxipParam)
	case "ipv4":
		ipParam, ok := params.(*packetparse.IPParams)
		if !ok {
			return "", errors.InvalidParam
		}
		return postflow.ApplyIPFlow(flows, url, *ipParam)
	case "ndn":
		ndnParam, ok := params.(*packetparse.NDNParams)
		if !ok {
			return "", errors.InvalidParam
		}
		return postflow.ApplyNdnFlow(flows, url, *ndnParam)
	case "geo":
		geoParam, ok := params.(*packetparse.GEOParams)
		if !ok {
			return "", errors.InvalidParam
		}
		return postflow.ApplyGEOFlow(flows, url, *geoParam)
	case "mf":
		mfParam, ok := params.(*packetparse.MFParams)
		if !ok {
			return "", errors.InvalidParam
		}
		return postflow.ApplyMfFlow(flows, url, *mfParam)
	case "id":
		idParam, ok := params.(*packetparse.IDParams)
		if !ok {
			return "", errors.InvalidParam
		}
		return postflow.ApplyIDFlow(flows, url, *idParam)
	default:
		return "invalid param", errors.InvalidParam
	}
}