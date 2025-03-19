package main

import (
	"onosutil/logic"

	"github.com/beego/beego/v2/server/web"
)

func setupRouter(manager *logic.Manager) {
	web.Post("/api/topo", manager.UpdateTopoHandler) // 更新拓扑结构
	web.Get("/api/topo", manager.GetTopoHandler)     // 查询拓扑结构

	web.Get("/api/tofino/port", manager.GetTofinoPortHandler) // 查询Tofino交换机模态对应转发端口
	web.Post("/api/checkpipe", manager.CheckPipeHandler)

	web.Router("/", &MainController{})
}
