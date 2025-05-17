package logic

import (
	"github.com/beego/beego/v2/server/web/context"
	log "github.com/sirupsen/logrus"
	"onosutil/utils/errors"
	"strings"
	"os/exec"
	"fmt"
)
func (m *Manager) SwitchSatelliteHandler(ctx *context.Context) {
	type SwitchSatelliteRequest struct {
		SwitchID  string  `json:"switch_id"`
	}
	var req SwitchSatelliteRequest
	if err := ctx.BindJSON(&req); err != nil {
		responseError(ctx, err)
		log.Errorf("SwitchSatelliteHandler参数绑定错误: %v", err)
		return
	}else{
		//解析到参数，执行武大提供的脚本 ssh luci@192.168.2.201 'sudo bash /home/luci/Desktop/ovsconf.sh  XXX'
		cmd:= exec.Command("ssh", "kin@10.190.96.114", "sudo bash /home/kin/Desktop/ovsconf.sh", fmt.Sprintf("%s", req.SwitchID))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("SwitchSatelliteHandler脚本执行错误: %v", err)
			responseError(ctx, err)
			return
		}
		log.Infof("SwitchSatelliteHandler脚本执行成功: %s", output)
		// 解析脚本输出，判断是否成功
		res := strings.TrimSpace(string(output))
		if res != "True" {
			log.Errorf("ModifyDevicePipeconfHandler failed: %s", res)
			responseError(ctx, errors.PipeconfCoverFailed)
			return
		}
		responseSuccess(ctx, res)
	}
}