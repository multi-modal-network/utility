package main

import (
	"onosutil/logic"

	"github.com/beego/beego/v2/server/web"
)

func setupRouter(manager *logic.Manager) {
	web.Post("/api/topo", manager.UpdateTopoHandler) // 更新拓扑结构
	web.Get("/api/topo", manager.GetTopoHandler)     // 查询拓扑结构

	web.Get("/api/tofino/port", manager.GetTofinoPortHandler)     // 查询Tofino交换机模态对应转发端口
	web.Post("/api/tofino/port", manager.ModifyTofinoPortHandler) // 修改Tofino交换机模态对应转发端口

	web.Post("/api/device/pipeconfs", manager.BatchCheckPipeconfHandler)  // 批量查询设备的pipeconf是否支持特定模态
	web.Get("/api/device/pipeconf", manager.GetDevicePipeconfHandler)     // 查询设备的pipeconf
	web.Post("/api/device/pipeconf", manager.ModifyDevicePipeconfHandler) // 修改设备的pipeconf（需要掉武大的流水线覆盖）

	web.Post("/api/traffic", manager.RecordTrafficHandler) // 处理流量三元组基础信息 更新路径
	web.Get("/api/traffic", manager.QueryTrafficHandler)   // 查询流量路径

	web.Router("/", &MainController{})
}
