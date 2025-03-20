package main

import (
	"onosutil/logic"

	"github.com/beego/beego/v2/server/web"
)

func setupRouter(manager *logic.Manager) {
	web.Post("/api/topo", manager.UpdateTopoHandler) // 更新拓扑结构
	web.Get("/api/topo", manager.GetTopoHandler)     // 查询拓扑结构

	web.Get("/api/tofino/port", manager.GetTofinoPortHandler) // 查询Tofino交换机模态对应转发端口
	web.Post("/api/checkpipe", manager.CheckPipeHandler)      // 批量查询设备的pipeconf是否支持特定模态

	web.Post("/api/traffic", manager.RecordTrafficHandler) // 处理流量三元组基础信息
	web.Get("/api/traffic", manager.QueryTrafficHandler)   // 处理流量转发路径信息

	web.Router("/", &MainController{})
}
