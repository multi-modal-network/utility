package main

import (
	"onosutil/logic"

	"github.com/beego/beego/v2/server/web"
)

func setupRouter(manager *logic.Manager) {
	web.Post("/api/topo", manager.UpdateTopoHandler)
	web.Get("/api/topo", manager.GetTopoHandler)
	web.Post("/api/checkpipe", manager.CheckPipeHandler)

	web.Router("/", &MainController{})
}
