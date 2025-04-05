package logic

import (
	log "github.com/sirupsen/logrus"
	"onosutil/model"
	"onosutil/utils/errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
)

// 输出结果到某个文件
func outputToFile(unsupported []string, modal string) {
	f, err := os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	cur := time.Now().Format("2006-01-02 15:04:05")
	for _, device := range unsupported {
		f.WriteString(cur + " " + modal + " is not supported by " + device + "\n")
	}
}

// GetDevicePipeconfHandler 获取设备对应的pipeconf信息
func (m *Manager) GetDevicePipeconfHandler(ctx *context.Context) {
	// 注意这里的Get参数定为了deviceID，实际指代的是deviceName
	deviceName := ctx.Input.Query("deviceID")
	if deviceName == "" {
		log.Errorf("GetDevicePipeconfHandler deviceID is empty")
		responseError(ctx, errors.InvalidParam)
		return
	}
	device := model.Device{}
	if err := m.db.QueryTable(model.Device{}).Filter("device_name__exact").One(&device); err != nil {
		log.Errorf("GetDevicePipeconfHandler query failed")
		responseError(ctx, err)
		return
	}
	res := DeviceInfo{
		ManagementAddress: device.ManagementAddress,
		Driver:            device.Driver,
		Pipeconf:          device.Pipeconf,
	}
	responseSuccess(ctx, res)
}

// ModifyDevicePipeconfHandler 修改设备的pipeconf（调用武大的流水线覆盖功能）
func (m *Manager) ModifyDevicePipeconfHandler(ctx *context.Context) {
	var req struct {
		DeviceID string `json:"deviceID"`
		Pipeconf string `json:"pipeconf"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		log.Errorf("ModifyDevicePipeconfHandler bindjson failed: %v", err)
		responseError(ctx, err)
		return
	}
	// 执行武大流水线覆盖程序（todo:确定程序路径）
	cmd := exec.Command("python3", "/home/onos/Desktop/pipeconf.py", "-d", req.DeviceID, "-p", req.Pipeconf)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("ModifyDevicePipeconfHandler exec failed: %v", err)
		responseError(ctx, err)
		return
	}
	res := strings.TrimSpace(string(output))
	if res != "True" {
		log.Errorf("ModifyDevicePipeconfHandler failed: %s", res)
		responseError(ctx, errors.PipeconfCoverFailed)
		return
	}
	responseSuccess(ctx, nil)
}
