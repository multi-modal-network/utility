package logic

import (
	"os/exec"

	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
)

func (m *Manager) SwitchSatelliteHandler(ctx *context.Context) {
	type SwitchSatelliteRequest struct {
		SwitchID string `json:"switch_id"`
	}
	var req SwitchSatelliteRequest
	if err := ctx.BindJSON(&req); err != nil {
		responseError(ctx, err)
		log.Errorf("SwitchSatelliteHandler参数绑定错误: %v", err)
		return
	} else {
		//解析到参数，执行武大提供的脚本ssh luci@192.168.2.201 'sudo bash /home/luci/Desktop/ovsconf.sh  XXX'
		cmd := exec.Command("sshpass", "-p", "Fiberhome@2020", "ssh", "-t", "onos4@10.190.96.114", "sudo", "/home/onos4/weixing/ovsconf.sh", req.SwitchID)
		// cmd := exec.Command("ssh", "onos4@10.190.96.114", "sudo", "bash", "/home/onos4/weixing/ovsconf.sh", req.SwitchID)
		//设置命令的输出
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("SwitchSatelliteHandler脚本执行错误: %v", err)
			responseError(ctx, err)
			return
		}
		log.Infof("SwitchSatelliteHandler脚本执行成功: %s", output)
		// 解析脚本输出，判断是否成功
		responseSuccess(ctx, nil)
	}
}
