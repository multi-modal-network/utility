package logic

import (
	"log"
	"onosutil/model"
	"os"
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

// 校正modelstring格式
func modelStringCorrect(modaltype string) string {
	switch modaltype {
	case "ipv4":
		return "IP"
	default:
		return strings.ToUpper(modaltype)
	}
}

func (m *Manager) CheckPipeHandler(ctx *context.Context) {
	var req struct {
		SendArray []string `json:"sendArray"`
		ModalType string   `json:"modalType"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		responseError(ctx, err)
		return
	}

	req.ModalType = modelStringCorrect(req.ModalType)

	unsupported := make([]string, 0)

	// 遍历sendArray，检查是否存在不支持的设备
	for _, device := range req.SendArray {
		//查找数据库中所有的表
		var res []*model.Device
		if _, err := m.db.QueryTable(&model.Device{}).Filter("device_id", device).All(&res); err != nil {
			responseError(ctx, err)
			return
		}

		if len(res) == 0 {
			log.Printf("device %s not found", device)
			continue
		}

		if strings.Contains(res[0].SupportModal, req.ModalType) {
			continue
		} else {
			unsupported = append(unsupported, device)
		}
	}

	type getUnsupportDeviceResponse struct {
		UnsupportDevices []string `json:"unsupported"`
	}
	//把结果输出到output
	outputToFile(unsupported, req.ModalType)
	// 返回结果
	responseSuccess(ctx, getUnsupportDeviceResponse{
		UnsupportDevices: unsupported,
	})

}
